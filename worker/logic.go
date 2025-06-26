package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sdcc-mapreduce/utils"
	"sort"
	"strings"
	"time"
)

// Worker gestisce i task di Map e Reduce
type Worker struct{}

// Esegue il task di Map: ordina il chunk di numeri ricevuto e lo invia ai reducer appropriati
func (Worker) MapTask(req utils.MapRequest, reply *utils.MapReply) error {
	log.Printf("Mapper ha ricevuto il chunk: %v\n", req.Chunk)
	time.Sleep(5 * time.Second)

	// Ordina i numeri localmente
	sort.Ints(req.Chunk)
	log.Printf("Chunk ordinato: %v\n", req.Chunk)

	// Scrive il chunk ordinato in un file temporaneo, stile reducer
	host, err := os.Hostname()
	if err != nil {
		host = "unknown_mapper"
	}
	tempFileName := fmt.Sprintf("output/temp_%s.txt", host)

	file, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Errore apertura %s: %v", tempFileName, err)
	} else {
		writer := bufio.NewWriter(file)
		for _, num := range req.Chunk {
			writer.WriteString(fmt.Sprintf("%d\n", num))
		}
		writer.Flush()
		file.Close()
		log.Printf("Mapper ha scritto i risultati in: %s\n", tempFileName)
	}


	// Mappa: reducer --> sotto-chunk assegnato
	assignments := make(map[string][]int)

	// Tutti i reducer disponibili
	var allReducers []string
	for addr := range req.ReducerRanges {
		allReducers = append(allReducers, addr)
	}

	// Assegna ciascun numero al reducer corretto
	for _, num := range req.Chunk {
		assigned := false
		for addr, bounds := range req.ReducerRanges {
			if num >= bounds[0] && num < bounds[1] {
				assignments[addr] = append(assignments[addr], num)
				assigned = true
				log.Printf("Numero %d --> %s (range %d-%d)", num, addr, bounds[0], bounds[1])
				break
			}
		}
		if !assigned {
			log.Printf("Errore: numero %d non assegnato a nessun reducer!\n", num)
		}
	}

	// Invia ogni sotto-chunk al reducer assegnato, con fallback se fallisce
	for primaryAddr, nums := range assignments {
		log.Printf("Invio il sotto-chunk %v al reducer %s (con fallback)\n", nums, primaryAddr)
		err := SendToReducerWithFallback(nums, primaryAddr, allReducers)
		if err != nil {
			log.Printf("Fallimento finale invio chunk %v: %v\n", nums, err)
		}
	}

	// ACK al master
	reply.Ack = true
	return nil
}

// Esegue task di Reduce
func (Worker) ReduceTask(req utils.ReduceRequest, reply *utils.ReduceReply) error {
  log.Printf("\nReducer ha ricevuto i seguenti chunk: %v per Owner %s\n", req.Chunks, req.Owner)

  // Usa Owner per formare il nome del file
  ownerSafe := strings.ReplaceAll(req.Owner, ":", "_")
  tempFileName := fmt.Sprintf("output/temp_%s.txt", ownerSafe)

  // Scrive nel file temporaneo
  file, err := os.OpenFile(tempFileName,
    os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    log.Fatalf("Errore apertura %s: %v", tempFileName, err)
  }
  defer file.Close()

  // Scrive ogni numero ricevuto in una nuova riga del file
  writer := bufio.NewWriter(file)
  for _, num := range req.Chunks {
    writer.WriteString(fmt.Sprintf("%d\n", num))
  }
  writer.Flush()

  log.Printf("Reducer ha scritto i risultati in: %s\n", tempFileName)

  // Invio ACK al master
  reply.Ack = true
  return nil
}


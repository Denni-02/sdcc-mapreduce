package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sdcc-mapreduce/utils"
	"sort"
	"strings"
)

// Worker gestisce i task di Map e Reduce
type Worker struct{}

// Esegue il task di Map: ordina il chunk di numeri ricevuto e lo invia ai reducer appropriati
func (Worker) MapTask(req utils.MapRequest, reply *utils.MapReply) error {
	numbers := req.Chunk               // chunk da elaborare
	reducerRanges := req.ReducerRanges // // Mappa {reducer address --> [min, max]}
	fmt.Printf("Mapper ha ricevuto il chunk: %v\n", numbers)

	// Ordina localmente il chunk
	sort.Ints(numbers)
	fmt.Printf("Chunk ordinato: %v\n", numbers)

	// Distribuisce i numeri ordinati ai reducer appropriati
	var v = make(map[string][]int) // Mappa {reducer --> []numeri da inviare}
	for _, num := range numbers {
		assigned := false
		for address, bounds := range reducerRanges {
			if num >= bounds[0] && num <= bounds[1] {
				v[address] = append(v[address], num)
				assigned = true
				break
			}
		}
		if !assigned {
			fmt.Printf("Errore: numero %d non assegnato a nessun reducer!\n", num)
		}
	}

	// Invia i sotto-chunk ordinati ai rispettivi reducer via RPC
	for address := range v {
		fmt.Printf("Invio il sotto-chunck %v al mapper %s\n", v[address], address)
		err := utils.SendToReducer(v[address], address)
		if err != nil {
			log.Printf("Errore nell'invio al reducer %s: %v\n", address, err)
			continue
		}
	}

	// Invia un ACK al master
	reply.Ack = true
	return nil
}

// Esegue task di Reduce
func (Worker) ReduceTask(req utils.ReduceRequest, reply *utils.ReduceReply) error {

	fmt.Printf("\nReducer ha ricevuto i seguenti chunk: %v\n", req.Chunks)

	// Nome univoco del file temporaneo basato sull'indirizzo del reducer
	workerAddress := flag.Lookup("address").Value.String()
	sanitizedAddress := strings.ReplaceAll(workerAddress, ":", "_")     // Sostituisce i ":" con "_"
	tempFileName := fmt.Sprintf("output/temp_%s.txt", sanitizedAddress) // Torna alla directory principale

	// Scrive nel file temporaneo
	file, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Errore nell'apertura del file temporaneo %s: %v", tempFileName, err)
	}
	fmt.Printf("Scrivo nel file temporaneo: %s\n", tempFileName)
	defer file.Close()

	// Scrive ogni numero ricevuto in una nuova riga del file
	writer := bufio.NewWriter(file)
	for _, num := range req.Chunks {
		writer.WriteString(fmt.Sprintf("%d\n", num))
	}
	writer.Flush()

	fmt.Printf("Reducer ha scritto i risultati nel file: %s\n", tempFileName)

	// Invio ACK al master
	reply.Ack = true
	return nil
}

func main() {
	// Flag per specificare indirizzo e porta dalla linea di comando
	address := flag.String("address", "localhost:9001", "Indirizzo e porta del worker (es. localhost:9001)")
	flag.Parse()

	worker := new(Worker)
	server := rpc.NewServer()
	err := server.Register(worker)
	if err != nil {
		log.Fatalf("Errore nella registrazione del worker: %v", err)
	}

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("Errore nell'ascolto su %s: %v", *address, err)
	}

	log.Printf("Worker in ascolto su %s\n", *address)
	server.Accept(listener)
}

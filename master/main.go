package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"
	"sdcc-mapreduce/utils"
)

// Coordina l'intero flusso MapReduce: generazione dati, assegnazione task, raccolta risultati.
func main() {

	// Pulizia dei file di output precedenti
	utils.CleanupOutputFiles()

	// Logger master
	logger, file, err := utils.SetupLogger("/app/log/log_master/master.log", "[MASTER] ")
	if err != nil {
		log.Fatalf("Errore logger master: %v", err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.SetFlags(logger.Flags())
	log.SetPrefix(logger.Prefix())

	// Carica la configurazione dal file config.json
	config := utils.LoadConfig("config/config.json")

	// Inizializza il master con i worker configurati
	master := Master{
		Workers:  config.Workers,
		Settings: config.Settings,
	}

	// Avvia il server RPC per la registrazione dei worker
	rpcServer := rpc.NewServer()
	err = rpcServer.Register(&master)
	if err != nil {
		log.Fatalf("Errore nella registrazione RPC del master: %v", err)
	}

	// Listener per accettare le connessioni dei worker
	go func() {
		listener, err := net.Listen("tcp", ":9000")
		if err != nil {
			log.Fatalf("Errore nell'ascolto su porta 9000: %v", err)
		}
		log.Println("Master RPC server in ascolto su :9000 per registrazioni")
		rpcServer.Accept(listener)
	}()

	// Aspetta che tutti i worker si registrino
	master.WaitForWorkers(config.Settings.NumMappers, config.Settings.NumReducers)

	// Genera i dati casuali
	data := master.GenerateData(config.Settings.Count, config.Settings.Xi, config.Settings.Xf)
	log.Printf("\n\nNumeri generati: %v\n\n", data)
	utils.SaveDataToFile(data)

	// Suddivide i dati in chunk
	chunks := master.SplitData(data)
	log.Printf("Chunk distribuiti ai mapper: %v\n\n", chunks)
	utils.SaveChunksToFile(chunks)

	// Mappa i reducer agli intervalli di competenza
	reducerRanges := master.MapReducersToRanges(data)

	fmt.Println("Tutti i worker registrati. Attendo 15 secondi prima della fase MAP per permettere fault injection...")
	time.Sleep(15 * time.Second)

	// Fase di Map
	master.ExecuteMapPhase(chunks, reducerRanges)

	// Combina i risultati finali
	master.CombineOutputFiles()
        os.Exit(0)
}

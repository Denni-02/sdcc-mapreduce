package main

import (
	"flag"
	"log"
	"fmt"
	"net"
	"net/rpc"
	"sdcc-mapreduce/utils"
)

func main() {

	// Recupera eventuali panic ed evita che il worker muoia silenziosamente
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic catturato: %v\n", r)
			utils.AppendToFile("/app/log/log_worker/worker_crash.log", fmt.Sprintf("Panic: %v\n", r))
		}
	}()

	// Flag per specificare indirizzo e porta dalla linea di comando
	address := flag.String("address", "localhost:9001", "Indirizzo e porta del worker (es. localhost:9001)")
	flag.Parse()

	// Inizializza logger: restituisce anche il file per il defer
	logFileName := "/app/log/log_worker/worker_" + utils.SanitizeAddr(*address) + ".log"
	logger, file, err := utils.SetupLogger(logFileName, "[WORKER] ")
	if err != nil {
		log.Fatalf("Errore inizializzazione logger: %v", err)
	}
	defer file.Close()
	log.SetOutput(file) 
	log.SetFlags(logger.Flags())      
	log.SetPrefix(logger.Prefix())

	// Crea una nuova istanza del worker che implementa i metodi RPC
	worker := new(Worker)
	server := rpc.NewServer()
	err = server.Register(worker)
	if err != nil {
		log.Fatalf("Errore nella registrazione del worker: %v", err)
	}

	// Crea un listener TCP sull'indirizzo specificato
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("Errore nell'ascolto su %s: %v", *address, err)
	}

	// Accetta le connessioni in arrivo e serve le richieste RPC
	log.Printf("Worker in ascolto su %s\n", *address)
	server.Accept(listener)
}


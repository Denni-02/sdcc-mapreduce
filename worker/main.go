package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
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

	// Legge variabili d’ambiente
	role := os.Getenv("ROLE")
	masterAddr := os.Getenv("MASTER_ADDR")
	if role == "" || masterAddr == "" {
		log.Fatalf("Variabili ROLE o MASTER_ADDR mancanti")
	}

	// Inizializza logger
	logFileName := "/app/log/log_worker/worker_" + utils.SanitizeAddr(*address) + ".log"
	logger, file, err := utils.SetupLogger(logFileName, "[WORKER] ")
	if err != nil {
		log.Fatalf("Errore inizializzazione logger: %v", err)
	}
	defer file.Close()
	log.SetOutput(file) 
	log.SetFlags(logger.Flags())      
	log.SetPrefix(logger.Prefix())

	// Invio di Register al master
	registerSelf(*address, role, masterAddr)

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

func registerSelf(address, role, masterAddr string) {
	client, err := rpc.Dial("tcp", masterAddr)
	if err != nil {
		log.Fatalf("Errore connessione RPC al master (%s): %v", masterAddr, err)
	}
	defer client.Close()

	req := utils.WorkerConfig{
		Role:    role,
		Address: address,
	}
	var reply bool

	err = client.Call("Master.Register", req, &reply)
	if err != nil {
		log.Fatalf("Errore RPC Register: %v", err)
	}
	if reply {
		log.Printf("Registrazione avvenuta con successo (%s - %s)", role, address)
	} else {
		log.Printf("Worker già registrato (%s)", address)
	}
}


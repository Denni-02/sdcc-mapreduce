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

// Tenta di inviare una RPC a uno dei mapper disponibili
func CallWithFallbackMapBusy(
	workers []utils.WorkerConfig,
	method string,
	request interface{},
	reply interface{},
	logPrefix string,
	taskLabel string,
	busy *utils.ThreadSafeMap,
) error {
	const maxRetries = 5
	const retryDelay = 2 * time.Second

	logFile, _ := os.OpenFile("/app/log/log_master/master.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	attempts := 0

	for retry := 1; retry <= maxRetries; retry++ {
		tried := make(map[string]bool)

		for _, worker := range workers {
			addr := worker.Address

			if tried[addr] || busy.Get(addr) {
				continue
			}

			tried[addr] = true
			attempts++
			busy.Set(addr, true)

			conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
			if err != nil {
				logger.Printf("[%s] Tentativo %d: connessione fallita a %s: %v", logPrefix, attempts, addr, err)
				busy.Set(addr, false)
				continue
			}

			client := rpc.NewClient(conn)
			err = client.Call(method, request, reply)
			client.Close()

			if err != nil {
				logger.Printf("[%s] Tentativo %d: errore RPC su %s: %v", logPrefix, attempts, addr, err)
				busy.Set(addr, false)
				continue
			}

			// Se la risposta è valida → successo
			if mapReply, ok := reply.(*utils.MapReply); ok && mapReply.Ack {
				busy.Set(addr, false)
				logger.Printf("[%s] Completato con successo da %s", logPrefix, addr)
				return nil
			}
		}

		// Nessun mapper disponibile in questo giro
		logger.Printf("[%s] Nessun mapper disponibile (tentativo %d/%d). Riprovo tra %v...", logPrefix, retry, maxRetries, retryDelay)
		time.Sleep(retryDelay)
	}

	// Fallimento definitivo dopo tutti i tentativi
	logger.Printf("[%s] Fallimento definitivo su tutti i mapper", logPrefix)
	appendToFile("/app/log/log_master/worker_failed_tasks.log", fmt.Sprintf("%s: %s\n", logPrefix, taskLabel))

	return fmt.Errorf("tutti i tentativi falliti per %s", taskLabel)
}


// CallWithRetry tenta di inviare una RPC a uno specifico reducer
func CallWithRetry(workerAddr string, method string, request interface{}, reply interface{}, logPrefix, taskLabel string) error {
	const maxRetries = 3

	logFile, _ := os.OpenFile("/app/log/log_master/master_failed_task.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		conn, err := net.DialTimeout("tcp", workerAddr, 3*time.Second)
		if err != nil {
			logger.Printf("[%s] Tentativo %d: connessione fallita (%s): %v", logPrefix, attempt, workerAddr, err)
			time.Sleep(1 * time.Second)
			continue
		}

		client := rpc.NewClient(conn)
		defer client.Close()

		err = client.Call(method, request, reply)
		if err != nil {
			logger.Printf("[%s] Tentativo %d: errore RPC %s: %v", logPrefix, attempt, method, err)
			time.Sleep(1 * time.Second)
			continue
		}

		logger.Printf("[%s] Successo da %s", logPrefix, workerAddr)
		return nil
	}

	// Dopo max tentativi si ha fallimento
	logger.Printf("[%s] Fallimento definitivo su %s", logPrefix, workerAddr)
	appendToFile("/app/log/log_master/failed_tasks.log", fmt.Sprintf("%s: %s\n", logPrefix, taskLabel))
	return fmt.Errorf("tutti i tentativi falliti per %s", taskLabel)
}

// Scrive una riga nel file specificato in modalità append
func appendToFile(path string, line string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Errore apertura %s: %v", path, err)
		return
	}
	defer f.Close()
	f.WriteString(line)
}

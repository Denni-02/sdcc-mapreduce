package main

import (
	"fmt"
	"net" 
	"log"
	"net/rpc"
	"sdcc-mapreduce/utils"
	"time"
)

// Invia un sotto-chunk a un reducer con retry automatico sullo stesso reducer
func SendToReducerWithRetry(nums []int, address string) error {
	const maxRetries = 3

	for i := 1; i <= maxRetries; i++ {
		err := utils.SendToReducer(nums, address)
		if err == nil {
			return nil
		}
		log.Printf("Tentativo %d fallito verso %s: %v\n", i, address, err)
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("tutti i tentativi falliti verso %s", address)
}


// Prova a inviare un sotto-chunk a un reducer primario e, in caso di errore, agli altri disponibili
func SendToReducerWithFallback(nums []int, primary string, allReducers []string) error {
	tried := make(map[string]bool) // Traccia i reducer già provati

	// Ordina i candidati: primario prima, poi tutti gli altri
	candidates := append([]string{primary}, allReducers...)

	for _, addr := range candidates {
		if tried[addr] {
			continue
		}
		tried[addr] = true

		// Prepara la richiesta RPC
		req := utils.ReduceRequest{
			Chunks:        nums,
			WorkerAddress: addr,
			Owner:         primary, 
		}
		var reply utils.ReduceReply

		// Connessione con timeout di 3 secondi
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err != nil {
			log.Printf("Fallito verso %s (timeout connessione): %v\n", addr, err)
			continue
		}

    // Invoca il metodo ReduceTask sul reducer
		client := rpc.NewClient(conn)
		err = client.Call("Worker.ReduceTask", req, &reply)
		client.Close()

		if err == nil && reply.Ack {
			log.Printf("Chunk inviato con successo a %s (primario %s)\n", addr, primary)
			return nil
		}
		log.Printf("Fallito verso %s (errore chiamata RPC): %v\n", addr, err)
	}

	// Se nessun reducer ha risposto con successo
	log.Printf("Tutti i fallback falliti per chunk: %v (primario %s)", nums, primary)
	utils.AppendToFile("/app/log/log_master/failed_tasks.log", fmt.Sprintf("Reduce fallito: chunk %v (--> %s)\n", nums, primary))
	return fmt.Errorf("Tutti i fallback falliti per chunk: %v (primario %s)", nums, primary)
}
package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"
	"os"
)

// Strutture per le chiamate RPC

// MapRequest e MapReply per la fase di Map
type MapRequest struct {
	Chunk         []int             // Dati del chunk da ordinare
	ReducerRanges map[string][2]int // Mappa dei range assegnati a ciascun reducer
}

type MapReply struct {
	Ack bool // Conferma dell'esecuzione da parte del mapper
}

// ReduceRequest e ReduceReply per la fase di Reduce
type ReduceRequest struct {
	Chunks        []int  `json:"chunks"`
	WorkerAddress string `json:"workerAddress"` // Aggiungi l'indirizzo del reducer
}

type ReduceReply struct {
	//SortedData []int // Dati ordinati restituiti dal reducer
	Ack bool
}

// Configurazione dei worker
type WorkerConfig struct {
	Role    string `json:"role"`    // Specifica il ruolo del worker (mapper/reducer)
	Address string `json:"address"` // Indirizzo del worker
}

// Configurazione generale del sistema
type Settings struct {
	NumMappers  int `json:"numMappers"`  // Numero di mapper
	NumReducers int `json:"numReducers"` // Numero di reducer
	Xi          int `json:"xi"`          // Valore minimo
	Xf          int `json:"xf"`          // Valore massimo
	Count       int `json:"count"`       // Numero di valori casuali generati
}

type Config struct {
	Workers  []WorkerConfig `json:"workers"`  // Lista dei worker
	Settings Settings       `json:"settings"` // Impostazioni generali del sistema
}

// Funzione per caricare la configurazione da un file JSON
func LoadConfig(path string) Config {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Errore nella lettura del file di configurazione: %v", err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Errore nel parsing del file di configurazione: %v", err)
	}

	return config
}

// Funzione per gestire errori
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Funzione per inviare i numeri ordinati al reducer appropriato
func SendToReducer(num []int, address string) error {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("errore di connessione al reducer %s: %v", address, err)
	}
	defer client.Close()

	req := ReduceRequest{
		Chunks: num,
	}
	reply := ReduceReply{}
	err = client.Call("Worker.ReduceTask", req, &reply)
	if err != nil {
		return fmt.Errorf("errore nell'esecuzione del ReduceTask: %v", err)
	}
	return nil
}

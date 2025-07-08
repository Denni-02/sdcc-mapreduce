package utils

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

var statusMu sync.Mutex

// SaveDataToFile salva i numeri generati in ./state/data.json e li carica su S3 se abilitato
func SaveDataToFile(data []int) {
	err := os.MkdirAll("state", os.ModePerm)
	if err != nil {
		log.Fatalf("Errore creazione cartella state/: %v", err)
	}

	filePath := "state/data.json"
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Errore creazione %s: %v", filePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		log.Fatalf("Errore scrittura JSON %s: %v", filePath, err)
	}

	log.Printf("[STATE] Dati salvati in %s", filePath)

	// Upload S3 se richiesto
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		s3Path := "s3://" + bucket + "/state/data.json"
		cmd := exec.Command("aws", "s3", "cp", filePath, s3Path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Errore upload su S3 (%s): %v\nOutput: %s", s3Path, err, string(output))
		} else {
			log.Printf("Upload su S3 riuscito: %s", s3Path)
		}
	}
}


// SaveChunksToFile salva i chunk generati in ./state/chunks.json e li carica su S3 se abilitato
func SaveChunksToFile(chunks [][]int) {
	err := os.MkdirAll("state", os.ModePerm)
	if err != nil {
		log.Fatalf("Errore creazione cartella state/: %v", err)
	}

	filePath := "state/chunks.json"
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Errore creazione %s: %v", filePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(chunks)
	if err != nil {
		log.Fatalf("Errore scrittura JSON %s: %v", filePath, err)
	}

	log.Printf("[STATE] Chunk salvati in %s", filePath)

	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		s3Path := "s3://" + bucket + "/state/chunks.json"
		cmd := exec.Command("aws", "s3", "cp", filePath, s3Path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Errore upload su S3 (%s): %v\nOutput: %s", s3Path, err, string(output))
		} else {
			log.Printf("Upload su S3 riuscito: %s", s3Path)
		}
	}
}

func InitStatusFile(nChunks int) {
	err := os.MkdirAll("state", os.ModePerm)
	if err != nil {
		log.Fatalf("Errore creazione cartella state/: %v", err)
	}

	status := make(map[string]string)
	for i := 0; i < nChunks; i++ {
		status[strconv.Itoa(i)] = "pending"
	}

	filePath := "state/status.json"
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Errore creazione %s: %v", filePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(status)
	if err != nil {
		log.Fatalf("Errore scrittura JSON %s: %v", filePath, err)
	}

	log.Printf("[STATE] Stato inizializzato in %s", filePath)

	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		s3Path := "s3://" + bucket + "/state/status.json"
		cmd := exec.Command("aws", "s3", "cp", filePath, s3Path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Errore upload status.json su S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Printf("Upload su S3 riuscito: %s", s3Path)
		}
	}
}

func SaveStatusAfterChunk(i int) {

	statusMu.Lock()
	defer statusMu.Unlock()
	
	filePath := "state/status.json"

	// Leggi lo stato attuale
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Errore apertura %s: %v", filePath, err)
	}
	defer file.Close()

	var status map[string]string
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&status)
	if err != nil {
		log.Fatalf("Errore decoding JSON %s: %v", filePath, err)
	}

	// Aggiorna lo stato
	chunkID := strconv.Itoa(i)
	status[chunkID] = "done"

	// Riscrivi il file
	fileOut, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Errore creazione %s: %v", filePath, err)
	}
	defer fileOut.Close()

	encoder := json.NewEncoder(fileOut)
	err = encoder.Encode(status)
	if err != nil {
		log.Fatalf("Errore scrittura JSON %s: %v", filePath, err)
	}

	log.Printf("[STATE] Stato aggiornato: chunk %d → done", i)

	// Upload su S3 se abilitato
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		s3Path := "s3://" + bucket + "/state/status.json"
		cmd := exec.Command("aws", "s3", "cp", filePath, s3Path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Errore upload status.json su S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Printf("Upload su S3 riuscito: %s", s3Path)
		}
	}
}

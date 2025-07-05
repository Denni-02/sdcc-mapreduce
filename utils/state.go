package utils

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
)

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


package utils

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"fmt"
)

var statusMu sync.Mutex

/* -------------------------------------------------------------
		ALTRO
-------------------------------------------------------------- */

// Salva un flag JSON che indica il completamento con successo dell’esecuzione
func SaveCompletionFlag() {
	os.MkdirAll("state", os.ModePerm)
	filePath := "state/completed.json"

	err := os.WriteFile(filePath, []byte(`{"completed": true}`), 0644)
	if err != nil {
		log.Printf("Errore salvataggio completed.json: %v", err)
	} else {
		log.Println("[STATE] Flag completamento salvato.")
	}

	// Upload su S3 se abilitato
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		s3Path := fmt.Sprintf("s3://%s/state/completed.json", bucket)
		cmd := exec.Command("aws", "s3", "cp", filePath, s3Path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Errore upload completed.json su S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Printf("Upload completed.json su S3 riuscito: %s", s3Path)
		}
	}
}

// Controlla se esiste il file di completamento
func CompletionFlagExists() bool {
	filePath := "state/completed.json"

	// Scarica da S3 se attivo
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		s3Path := fmt.Sprintf("s3://%s/state/completed.json", bucket)
		cmd := exec.Command("aws", "s3", "cp", s3Path, filePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Errore download completed.json da S3: %v\nOutput: %s", err, string(output))
		}
		log.Println("[STATE] Scaricato completed.json da S3")
	}

	_, err := os.Stat(filePath)
	return err == nil
}

// Elimina il file di completamento
func RemoveCompletionFlag() {
	filePath := "state/completed.json"

	// Rimozione locale
	if err := os.Remove(filePath); err == nil {
		log.Println("[STATE] completed.json rimosso")
	} else if !os.IsNotExist(err) {
		log.Printf("Errore rimozione completed.json: %v", err)
	}

	// Rimozione da S3
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		s3Path := fmt.Sprintf("s3://%s/state/completed.json", bucket)
		cmd := exec.Command("aws", "s3", "rm", s3Path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Errore rimozione completed.json da S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Printf("completed.json rimosso da S3: %s", s3Path)
		}
	}
}

// Elimina tutti i file di stato
func ResetState() {
	// Elenco dei file locali da rimuovere
	files := []string{
		"state/status.json",
		"state/data.json",
		"state/chunks.json",
		"state/workers.json", 
	}

	for _, filePath := range files {
		if err := os.Remove(filePath); err == nil {
			log.Printf("[STATE] File %s rimosso", filePath)
		} else if !os.IsNotExist(err) {
			log.Printf("Errore durante la rimozione di %s: %v", filePath, err)
		}
	}

	// Rimuovi da S3 se attivo
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		for _, name := range []string{"status.json", "data.json", "chunks.json", "workers.json"} {
			s3Path := fmt.Sprintf("s3://%s/state/%s", bucket, name)
			cmd := exec.Command("aws", "s3", "rm", s3Path)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Errore rimozione %s da S3: %v\nOutput: %s", s3Path, err, string(output))
			} else {
				log.Printf("File rimosso da S3: %s", s3Path)
			}
		}
	}
}

// Salva i numeri generati in ./state/data.json e li carica su S3 se abilitato
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

// Salva i chunk generati in ./state/chunks.json e li carica su S3 se abilitato
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

// Inizializza il file status.json con lo stato "pending" per ogni chunk
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

// Aggiorna lo stato del chunk i-esimo a "done" in status.json
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

// Controlla la presenza del file status.json
func StateFilesExist() bool {
	_, err := os.Stat("state/status.json")
	return err == nil
}

/* -------------------------------------------------------------
		(1) CRASH PRIMA/DOPO REGISTRAZIONE WORKER
	-------------------------------------------------------------- */

// Verifica se esiste workers.json
func WorkersFileExists() bool {
	_, err := os.Stat("state/workers.json")
	return err == nil
}

// Legge i worker salvati in workers.json e li restituisce
func RecoverWorkersFromFile() []WorkerConfig {
	file, err := os.Open("state/workers.json")
	if err != nil {
		log.Printf("Errore apertura workers.json: %v", err)
		return nil
	}
	defer file.Close()

	var workers []WorkerConfig
	err = json.NewDecoder(file).Decode(&workers)
	if err != nil {
		log.Printf("Errore decoding workers.json: %v", err)
		return nil
	}
	return workers
}

// Salva i worker registrati in workers.json
func SaveWorkerOnRegister(workers []WorkerConfig) {
	err := os.MkdirAll("state", os.ModePerm)
	if err != nil {
		log.Printf("Errore creazione cartella state/: %v", err)
		return
	}

	filePath := "state/workers.json"
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Errore creazione workers.json: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(workers); err != nil {
		log.Printf("Errore encoding workers.json: %v", err)
		return
	}

	log.Println("[STATE] workers.json salvato correttamente.")

	// Upload S3 se abilitato
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		s3Path := fmt.Sprintf("s3://%s/state/workers.json", bucket)
		cmd := exec.Command("aws", "s3", "cp", filePath, s3Path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Errore upload workers.json su S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Printf("Upload workers.json su S3 riuscito: %s", s3Path)
		}
	}
}

/* -------------------------------------------------------------
		(2) FASE MAP GIA' COMPLETATA
-------------------------------------------------------------- */

// Controlla se tutti i chunk nel file status.json sono "done"
func PhaseAlreadyDone() bool {
	filePath := "state/status.json"

	// Se S3 abilitato, scarica la versione aggiornata
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		s3Path := "s3://" + bucket + "/state/status.json"
		cmd := exec.Command("aws", "s3", "cp", s3Path, filePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Errore download status.json da S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Printf("[STATE] Scaricato status.json aggiornato da S3")
		}
	}

	if !StateFilesExist() {
		log.Println("[STATE] Nessun file di stato trovato: skip PhaseAlreadyDone.")
		return false
	}

	// Apre il file locale
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Errore apertura %s: %v", filePath, err)
		return false
	}
	defer file.Close()

	// Decodifica
	var status map[string]string
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&status)
	if err != nil {
		log.Printf("Errore decoding JSON %s: %v", filePath, err)
		return false
	}

	// Controlla se tutti sono "done"
	for _, v := range status {
		if v != "done" {
			return false
		}
	}
	return true
}

/* -------------------------------------------------------------
		(3) FASE MAP INIZIATA E CHUNK PENDENTI
-------------------------------------------------------------- */

// Restituisce tutti i chunk marcati come "pending" nel file status.json
func RecoverPendingChunks() [][]int {
	statusMu.Lock()
	defer statusMu.Unlock()

	statusPath := "state/status.json"
	chunksPath := "state/chunks.json"

	// Scarica da S3 se abilitato
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")

		// Prova a scaricare status.json
		cmd := exec.Command("aws", "s3", "cp", fmt.Sprintf("s3://%s/state/status.json", bucket), statusPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("[RECOVERY] Warning: status.json non trovato su S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Println("[RECOVERY] status.json scaricato da S3")
		}

		// Prova a scaricare chunks.json
		cmd = exec.Command("aws", "s3", "cp", fmt.Sprintf("s3://%s/state/chunks.json", bucket), chunksPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("[RECOVERY] Warning: chunks.json non trovato su S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Println("[RECOVERY] chunks.json scaricato da S3")
		}
	}

	// Verifica se i file esistono localmente
	if _, err := os.Stat(statusPath); os.IsNotExist(err) {
		log.Println("[RECOVERY] Stato assente: nessun chunk pending da recuperare.")
		return [][]int{}
	}

	if _, err := os.Stat(chunksPath); os.IsNotExist(err) {
		log.Println("[RECOVERY] Errore: chunks.json assente. Recovery impossibile.")
		return [][]int{}
	}

	// Decodifica status.json
	statusFile, err := os.Open(statusPath)
	if err != nil {
		log.Printf("[RECOVERY] Errore apertura %s: %v", statusPath, err)
		return [][]int{}
	}
	defer statusFile.Close()

	var status map[string]string
	if err := json.NewDecoder(statusFile).Decode(&status); err != nil {
		log.Printf("[RECOVERY] Errore decoding status.json: %v", err)
		return [][]int{}
	}

	// Decodifica chunks.json
	chunksFile, err := os.Open(chunksPath)
	if err != nil {
		log.Printf("[RECOVERY] Errore apertura %s: %v", chunksPath, err)
		return [][]int{}
	}
	defer chunksFile.Close()

	var chunks [][]int
	if err := json.NewDecoder(chunksFile).Decode(&chunks); err != nil {
		log.Printf("[RECOVERY] Errore decoding chunks.json: %v", err)
		return [][]int{}
	}

	// Estrai i pending
	var pending [][]int
	for i, chunk := range chunks {
		if status[strconv.Itoa(i)] == "pending" {
			pending = append(pending, chunk)
		}
	}

	log.Printf("[RECOVERY] Trovati %d chunk pending", len(pending))
	return pending
}

/* -------------------------------------------------------------
		(4) CRASH DOPO GENERAZIONE DATI
-------------------------------------------------------------- */

// Controlla se il file data.json esiste
func DataFileExists() bool {
	_, err := os.Stat("state/data.json")
	return err == nil
}

// Carica i dati da data.json
func LoadDataFromFile() []int {
	filePath := "state/data.json"

	// Scarica da S3 se abilitato
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		cmd := exec.Command("aws", "s3", "cp", fmt.Sprintf("s3://%s/state/data.json", bucket), filePath)
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("[RECOVERY] Warning: errore download data.json da S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Println("[RECOVERY] data.json scaricato da S3")
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("[RECOVERY] data.json non esistente. Ritorno vuoto.")
		return []int{}
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Errore apertura %s: %v", filePath, err)
		return []int{}
	}
	defer file.Close()

	var data []int
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		log.Printf("Errore decoding JSON %s: %v", filePath, err)
		return []int{}
	}
	return data
}

/* -------------------------------------------------------------
		(5) CRASH DOPO GENERAZIONE CHUNK
-------------------------------------------------------------- */

// Controlla la presenza di chunks.json
func ChunkFileExists() bool {
	_, err := os.Stat("state/chunks.json")
	return err == nil
}

// Carica i chunk da chunks.json
func LoadChunksFromFile() [][]int {
	filePath := "state/chunks.json"

	// Scarica da S3 se abilitato
	if os.Getenv("ENABLE_S3") == "true" {
		bucket := os.Getenv("S3_BUCKET")
		cmd := exec.Command("aws", "s3", "cp", fmt.Sprintf("s3://%s/state/chunks.json", bucket), filePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("[RECOVERY] Warning: errore download chunks.json da S3: %v\nOutput: %s", err, string(output))
		} else {
			log.Println("[RECOVERY] chunks.json scaricato da S3")
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Errore apertura %s: %v", filePath, err)
		return [][]int{}
	}
	defer file.Close()

	var chunks [][]int
	if err := json.NewDecoder(file).Decode(&chunks); err != nil {
		log.Printf("Errore decoding chunks.json: %v", err)
		return [][]int{}
	}
	return chunks
}


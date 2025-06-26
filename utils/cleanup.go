package utils

import (
	"log"
	"os"
	"path/filepath"
)

// Rimuove tutti i file di output temporanei e finali
func CleanupOutputFiles() {
	finalFile := "output/final_output.txt"
	if err := os.Remove(finalFile); err == nil {
		log.Printf("File %s rimosso.\n", finalFile)
	} else if !os.IsNotExist(err) {
		log.Printf("Errore nella rimozione del file %s: %v\n", finalFile, err)
	}

	tempFiles, err := filepath.Glob("output/temp_*.txt")
	if err != nil {
		log.Fatalf("Errore nella ricerca dei file temporanei: %v", err)
	}

	for _, tempFile := range tempFiles {
		if err := os.Remove(tempFile); err == nil {
			log.Printf("File temporaneo %s rimosso.\n", tempFile)
		} else {
			log.Printf("Errore nella rimozione del file temporaneo %s: %v\n", tempFile, err)
		}
	}
}

// Elimina tutti i file presenti in log/, ma non la cartella.
func ClearOutputDir() {
	err := os.MkdirAll("log", os.ModePerm) // assicurati che esista
	if err != nil {
		log.Fatalf("Errore nella creazione della cartella log/: %v", err)
	}

	files, err := filepath.Glob("log/*")
	if err != nil {
		log.Fatalf("Errore nel trovare i file in output/: %v", err)
	}

	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			log.Printf("Errore nella rimozione di %s: %v", file, err)
		} else {
			log.Printf("File %s eliminato.", file)
		}
	}
}
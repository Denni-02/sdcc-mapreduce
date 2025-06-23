package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func CleanupFiles() {
	finalFile := "output/final_output.txt"
	if err := os.Remove(finalFile); err == nil {
		fmt.Printf("File %s rimosso.\n", finalFile)
	} else if !os.IsNotExist(err) {
		log.Printf("Errore nella rimozione del file %s: %v\n", finalFile, err)
	}

	tempFiles, err := filepath.Glob("output/temp_*.txt")
	if err != nil {
		log.Fatalf("Errore nella ricerca dei file temporanei: %v", err)
	}

	for _, tempFile := range tempFiles {
		if err := os.Remove(tempFile); err == nil {
			fmt.Printf("File temporaneo %s rimosso.\n", tempFile)
		} else {
			log.Printf("Errore nella rimozione del file temporaneo %s: %v\n", tempFile, err)
		}
	}
}

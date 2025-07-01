package utils

import (
	"log"
	"os"
	"strings"
)

// Reindirizza l'output del logger verso un file specifico
func SetupLogger(filepath string, prefix string) (*log.Logger, *os.File, error) {
	// Estrai solo la cartella (es. log_master o log_worker)
	dir := filepath[:strings.LastIndex(filepath, "/")]
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	logger := log.New(file, prefix, log.LstdFlags)
	return logger, file, nil
}


// SanitizeAddr rimuove ":" da un indirizzo (es. localhost:9001 → localhost_9001)
func SanitizeAddr(addr string) string {
	return strings.ReplaceAll(addr, ":", "_")
}

// Scrive una riga nel file specificato in modalità append
func AppendToFile(path string, line string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Errore apertura %s: %v", path, err)
		return
	}
	defer f.Close()
	f.WriteString(line)
}
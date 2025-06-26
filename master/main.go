package main

import (
	"log"
	"sdcc-mapreduce/utils"
)

// Coordina l'intero flusso MapReduce: generazione dati, assegnazione task, raccolta risultati.
func main() {

	// Logger master
	logger, file, err := utils.SetupLogger("/app/log/log_master/master.log", "[MASTER] ")
	if err != nil {
		log.Fatalf("Errore logger master: %v", err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.SetFlags(logger.Flags())
	log.SetPrefix(logger.Prefix())


	// Pulizia dei file di output precedenti
	utils.ClearOutputDir()
	utils.CleanupOutputFiles()

	// Carica la configurazione dal file config.json
	config := utils.LoadConfig("config/config.json")

	// Inizializza il master con i worker configurati
	master := Master{
		Workers:  config.Workers,
		Settings: config.Settings,
	}

	// Genera i dati casuali
	data := master.GenerateData(config.Settings.Count, config.Settings.Xi, config.Settings.Xf)
	log.Printf("\n\nNumeri generati: %v\n\n", data)

	// Suddivide i dati in chunk
	chunks := master.SplitData(data)
	log.Printf("Chunk distribuiti ai mapper: %v\n\n", chunks)

	// Mappa i reducer agli intervalli di competenza
	reducerRanges := master.MapReducersToRanges(data)

	// Fase di Map
	master.ExecuteMapPhase(chunks, reducerRanges)

	// Combina i risultati finali
	master.CombineOutputFiles()
}

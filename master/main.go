package main

import (
	"log"
	"net"
	"net/rpc"
	"time"
	"fmt"
	"sdcc-mapreduce/utils"
)

// Coordina l'intero flusso MapReduce: generazione dati, assegnazione task, raccolta risultati.
func main() {

	/* -------------------------------------------------------------
		(0) INIZIALIZZAZIONE
	-------------------------------------------------------------- */ 

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
	if utils.CompletionFlagExists() {
		utils.CleanupOutputFiles()
		utils.RemoveCompletionFlag()
	}

	// Carica la configurazione dal file config.json
	config := utils.LoadConfig("config/config.json")

	// Inizializza il master con i worker configurati
	master := Master{
		Workers:  config.Workers,
		Settings: config.Settings,
	}

	// Avvia il server RPC per la registrazione dei worker
	rpcServer := rpc.NewServer()
	err = rpcServer.Register(&master)
	if err != nil {
		log.Fatalf("Errore nella registrazione RPC del master: %v", err)
	}

	fmt.Println("[TEST1] Pausa per kill del master prima della registrazione")
	time.Sleep(15 * time.Second)

	// Listener per accettare le connessioni dei worker
	go func() {
		listener, err := net.Listen("tcp", ":9000")
		if err != nil {
			log.Fatalf("Errore nell'ascolto su porta 9000: %v", err)
		}
		log.Println("Master RPC server in ascolto su :9000 per registrazioni")
		rpcServer.Accept(listener)
	}()

	/* -------------------------------------------------------------
		(1) CRASH PRIMA/DOPO REGISTRAZIONE WORKER
	-------------------------------------------------------------- */

	// Recupera worker da file se esiste
	if utils.WorkersFileExists() {
		log.Println("[RECOVERY] Trovato workers.json → recupero worker registrati")
		master.Workers = utils.RecoverWorkersFromFile()
	}

	/* -------------------------------------------------------------
		(2) FASE MAP GIA' COMPLETATA
	-------------------------------------------------------------- */

	if utils.PhaseAlreadyDone() {
		log.Println("MAP già completata. Passo al Combine.")
		master.CombineOutputFiles()
		utils.ResetState()
		return
	}

	/* -------------------------------------------------------------
		(3) CRASH CON FASE MAP INIZIATA E CHUNK PENDENTI
	-------------------------------------------------------------- */
	if utils.StateFilesExist() {
		log.Println("[STATE] status.json esiste. Provo a recuperare i chunk pending...")

		data := utils.LoadDataFromFile()
		chunks := utils.RecoverPendingChunks()

		if len(chunks) > 0 {
			log.Printf("[RECOVERY] Avvio fase MAP con %d chunk pending\n", len(chunks))
			reducerRanges := master.MapReducersToRanges(data)
			master.ExecuteMapPhase(chunks, reducerRanges)
			master.CombineOutputFiles()
			utils.SaveCompletionFlag()
			utils.ResetState()
			return
		}

		log.Println("[RECOVERY] Nessun chunk pending trovato. Passo a generazione nuova.")
	}

	/* -------------------------------------------------------------
		(4) CRASH DOPO GENERAZIONE DATI
	-------------------------------------------------------------- */

	if utils.DataFileExists() {
		log.Println("[RECOVERY] Trovato solo data.json. Rilancio split.")
		data := utils.LoadDataFromFile()
		chunks := master.SplitData(data)
		utils.SaveChunksToFile(chunks)
		utils.InitStatusFile(len(chunks))
		reducerRanges := master.MapReducersToRanges(data)
		master.ExecuteMapPhase(chunks, reducerRanges)
		master.CombineOutputFiles()
		utils.SaveCompletionFlag()
		utils.ResetState()
		return
	}

	/* -------------------------------------------------------------
		(5) CRASH DOPO GENERAZIONE CHUNK
	-------------------------------------------------------------- */
	if utils.ChunkFileExists() {
		log.Println("[RECOVERY] Trovato solo chunk.json.")
		data := utils.LoadDataFromFile()
		chunks := utils.LoadChunksFromFile()
		utils.SaveChunksToFile(chunks)
		utils.InitStatusFile(len(chunks))
		reducerRanges := master.MapReducersToRanges(data)
		master.ExecuteMapPhase(chunks, reducerRanges)
		master.CombineOutputFiles()
		utils.SaveCompletionFlag()
		utils.ResetState()
		return
	}

	/* -------------------------------------------------------------
		(6) GENERAZIONE DA ZERO
	-------------------------------------------------------------- */

	master.WaitForWorkers(config.Settings.NumMappers, config.Settings.NumReducers)
	utils.SaveWorkerOnRegister(master.Workers)

	//fmt.Println("[TEST2] Pausa per kill del master dopo la registrazione ma prima della generazione dei dati")
	//time.Sleep(15 * time.Second)

	data := master.GenerateData(config.Settings.Count, config.Settings.Xi, config.Settings.Xf)
	utils.SaveDataToFile(data)

	//fmt.Println("[TEST3] Pausa per kill del master dopo generazione dati ma prima dello split in chunk")
	//time.Sleep(15 * time.Second)

	chunks := master.SplitData(data)
	utils.SaveChunksToFile(chunks)
	utils.InitStatusFile(len(chunks))

	reducerRanges := master.MapReducersToRanges(data)

	//fmt.Println("[TEST4] Pausa per kill del master prima di ExecuteMapPhase")
	//time.Sleep(15 * time.Second)

	master.ExecuteMapPhase(chunks, reducerRanges)

	//fmt.Println("[TEST5] Pausa per kill del master dopo di ExecuteMapPhase")
	//time.Sleep(15 * time.Second)

	master.CombineOutputFiles()
	utils.SaveCompletionFlag()
	utils.ResetState()
}

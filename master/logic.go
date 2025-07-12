package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sdcc-mapreduce/utils"
	"sort"
	"strings"
	"sync"
	"time"
)

// ========================================================================================
// Struct Master 
// ========================================================================================

// Master rappresenta il nodo centrale che coordina le fasi di Map e Reduce
type Master struct {
	Workers  []utils.WorkerConfig // Lista dei worker (mappers e reducers)
	Settings utils.Settings       // Parametri generali del sistema
	mu       sync.Mutex  // per accesso concorrente a workers
}

// ========================================================================================
// Recupero Mapper e Reducer da config
// ========================================================================================

// Recupera solo i mapper
func (m *Master) getMappers() (mappers []utils.WorkerConfig, numMappers int) {
	numMappers = 0
	for _, worker := range m.Workers {
		if worker.Role == "mapper" {
			mappers = append(mappers, worker)
			numMappers++
		}
	}
	return
}

// Recupera solo i reducer
func (m *Master) getReducers() (reducers []utils.WorkerConfig, numReducers int) {
	numReducers = 0
	for _, worker := range m.Workers {
		if worker.Role == "reducer" {
			reducers = append(reducers, worker)
			numReducers++
		}
	}
	return
}

// ========================================================================================
// Generazione, Suddivisione e Sampling dati
// ========================================================================================

// Genera count numeri casuali nel range [xi, xf]
func (m *Master) GenerateData(count, xi, xf int) []int {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	data := make([]int, count)
	for i := range data {
		data[i] = random.Intn(xf-xi+1) + xi
	}
	return data
}

// Divide la lista dei numeri in numMappers chunk, da assegnare ai mapper
func (m *Master) SplitData(data []int) [][]int {
	_, numChunks := m.getMappers()
	chunkSize := int(math.Ceil(float64(len(data)) / float64(numChunks)))
	chunks := make([][]int, 0)
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

// Rimescola casualmente la slice fornita
func shuffle(slice []int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // Crea un'istanza di rand con un seed
	r.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// ========================================================================================
// Logica e Fase MAP
// ========================================================================================

// Metodo RPC chiamato dai worker per registrarsi al master
func (m *Master) Register(worker utils.WorkerConfig, reply *bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Evita duplicati
	for _, w := range m.Workers {
		if w.Address == worker.Address {
			*reply = false
			return nil
		}
	}

	m.Workers = append(m.Workers, worker)
	log.Printf("Registrato nuovo worker: %s (%s)\n", worker.Address, worker.Role)
	*reply = true
	return nil
}

// Attende che si registrino tutti i worker richiesti (mapper + reducer)
func (m *Master) WaitForWorkers(expectedMappers, expectedReducers int) {
	const timeoutSec = 30
	const checkInterval = 1 * time.Second

	log.Printf("Attendo la registrazione di %d mapper e %d reducer...\n", expectedMappers, expectedReducers)
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)

	for {
		m.mu.Lock()
		mappers := 0
		reducers := 0
		for _, w := range m.Workers {
			if w.Role == "mapper" {
				mappers++
			} else if w.Role == "reducer" {
				reducers++
			}
		}
		m.mu.Unlock()

		log.Printf("Registrati finora: %d mapper, %d reducer\n", mappers, reducers)

		if mappers >= expectedMappers && reducers >= expectedReducers {
			log.Println("Tutti i worker sono registrati, si può partire.")
			break
		}

		if time.Now().After(deadline) {
			log.Printf("Timeout: non tutti i worker si sono registrati in %d secondi", timeoutSec)
			os.Exit(1)
		}

		time.Sleep(checkInterval)
	}
}

// Assegna a ciascun reducer un intervallo [min, max] di valori e usa sampling per definire range bilanciati
func (m *Master) MapReducersToRanges(data []int) map[string][2]int {
	reducerRanges := make(map[string][2]int)
	reducers, numReducers := m.getReducers()

	// Shuffle per avere un sample casuale
	shuffle(data)

	// Sample: minimo 10% del dataset, massimo tutto
	sampleSize := int(len(data) / 10)
	if sampleSize < numReducers {
		sampleSize = numReducers
	}
	if sampleSize > len(data) {
		sampleSize = len(data)
	}
	sample := data[:sampleSize]
	log.Printf("sample = %v\n", sample)

	// Se troppi reducer rispetto al sample, li riduciamo
	if numReducers > len(sample) {
		log.Printf("Troppi reducer per il sample (%d reducer, %d sample). Uso solo %d reducer.\n", numReducers, len(sample), len(sample))
		numReducers = len(sample)
		// anche i reducer da usare vanno tagliati
		reducers = reducers[:numReducers]
	}

	// Calcolo range bilanciati
	ranges := createBalancedRanges(sample, numReducers)
	log.Printf("ranges = %v\n", ranges)

	// Protezione in caso ranges sia più corto
	if len(ranges) < numReducers-1 {
		log.Printf("Ranges troppo corti (%d), riduco numReducers a %d", len(ranges), len(ranges)+1)
		numReducers = len(ranges) + 1
		reducers = reducers[:numReducers]
	}

	for i, reducer := range reducers {
		var lower, upper int
		if i == 0 {
			lower = m.Settings.Xi
			upper = ranges[0]
		} else if i == numReducers-1 {
			lower = ranges[i-1]
			upper = m.Settings.Xf + 1
		} else {
			lower = ranges[i-1]
			upper = ranges[i]
		}
		reducerRanges[reducer.Address] = [2]int{lower, upper}
		log.Printf("Reducer %s gestisce l'intervallo [%d, %d]\n", reducer.Address, lower, upper)
	}

	log.Printf("Reducer effettivamente utilizzati: %d\n", numReducers)
	return reducerRanges
}


// Ordina il sample e ritorna N−1 punti di taglio per N reducer
func createBalancedRanges(sample []int, numReducers int) []int {
	sort.Ints(sample)
	log.Printf("sample ordinato= %v\n", sample)

	ranges := make([]int, 0)
	step := len(sample) / numReducers

	for i := 1; i < numReducers; i++ {
		index := i * step
		if index >= len(sample) {
			break
		}
		val := sample[index]
		// Evita valori duplicati consecutivi
		if len(ranges) == 0 || val != ranges[len(ranges)-1] {
			ranges = append(ranges, val)
		}
	}

	log.Printf("ranges= %v\n", ranges)
	return ranges
}


// Esegue la fase di Map assegnando ogni chunk di dati a un mapper disponibile
func (m *Master) ExecuteMapPhase(chunks [][]int, reducerRanges map[string][2]int) {
	
	var wg sync.WaitGroup // WaitGroup per sincronizzare le goroutine
	mappers, _ := m.getMappers() // Recupera la lista dei mapper dal maste

	// Mappa per tenere traccia dei mapper occupati con mutex per accesso concorrente
	busy := make(map[string]bool)
	var mu sync.Mutex

	for i, chunk := range chunks {
		wg.Add(1)
		go func(chunkIndex int, chunk []int) {
			defer wg.Done()

			req := utils.MapRequest{
				Chunk: chunk, // Chunk di interi da ordinare
				ReducerRanges: reducerRanges,  // Intervalli di valori per ogni reducer
			}
			reply := utils.MapReply{} // Struttura di risposta RPC

			logPrefix := fmt.Sprintf("MAP-%02d", chunkIndex)
			taskLabel := fmt.Sprintf("chunk %d --> %v", chunkIndex, chunk)
			
			// Chiamata RPC con fallback: tenta mapper disponibili e riassegna in caso di fallimento
			busyMap := &utils.ThreadSafeMap{Data: busy, Mu: &mu}
			err := CallWithFallbackMapBusy(mappers, "Worker.MapTask", req, &reply, logPrefix, taskLabel, busyMap)

			if err != nil {
				log.Printf("%s fallito: %v\n", logPrefix, err)
			} else {
				log.Printf("%s completato\n", logPrefix)
				utils.SaveStatusAfterChunk(chunkIndex)
			}
		}(i, chunk)
	}
	wg.Wait()
	log.Println("Fase di Map completata.")
}

// ========================================================================================
// Combinazione output
// ========================================================================================

// Combina i file di output dei reducer
func (m *Master) CombineOutputFiles() {
	outputFile := "output/final_output.txt"
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Errore nella creazione del file di output: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	reducers, _ := m.getReducers()

	for _, reducer := range reducers {
		tempFile := fmt.Sprintf("output/temp_%s.txt", strings.ReplaceAll(reducer.Address, ":", "_"))
		log.Printf("Unisco il file temporaneo: %s\n", tempFile)

		// Prova ad aprire il file, se non esiste logga e salta
		content, err := os.ReadFile(tempFile)
		if err != nil {
			log.Printf("File %s non trovato o vuoto. Skippato.\n", tempFile)
			continue
		}

		// Scrive il contenuto
		writer.Write(content)
		writer.WriteString("\n")
	}

	writer.Flush()
	log.Printf("Output finale scritto in: %s\n", outputFile)
}

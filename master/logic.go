package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/rpc"
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

// Assegna a ciascun reducer un intervallo [min, max] di valori e usa sampling per definire range bilanciati
func (m *Master) MapReducersToRanges(data []int) map[string][2]int {

	// Mappa per salvare il range per ogni reducer
	reducerRanges := make(map[string][2]int)
	reducers, numReducers := m.getReducers()

	// Shuffle dei dati
	shuffle(data)

	// Calcolo dimensione sample (10% dataset)
	tenPercentLenght := int(len(data) / 10)

	// Estrae un sample rappresentativo dei dati
	var sample []int
	if tenPercentLenght >= numReducers {
		sample = data[0:tenPercentLenght]
	} else {
		sample = data[0:numReducers] // se ha pochi dati almeno un valore per reducer
	}
	fmt.Printf("sample = %v\n", sample)

	// Crea intervalli bilanciat in base al sample
	ranges := createBalancedRanges(sample, numReducers)
	fmt.Printf("ranges = %v\n", ranges)

	for i, reducer := range reducers {
		var lower int
		var upper int

		if i == 0 { // Primo reducer: da xi al primo range
			lower = m.Settings.Xi
			upper = ranges[0]
		} else if i == numReducers-1 { // Ultimo reducer: dall'ultimo range fino a xf
			upper = m.Settings.Xf + 1
			lower = ranges[i-1]
		} else { // Reducer intermedi
			lower = ranges[i-1]
			upper = ranges[i]
		}

		reducerRanges[reducer.Address] = [2]int{lower, upper}
		fmt.Printf("Reducer %s gestisce l'intervallo [%d, %d]\n", reducer.Address, lower, upper)
	}
	return reducerRanges
}

// Ordina il sample e ritorna N−1 punti di taglio per N reducer
func createBalancedRanges(sample []int, numReducers int) []int {

	// Ordina il sample, così puoi dividerlo in ordine crescente
	sort.Ints(sample)
	fmt.Printf("sample ordinato= %v\n", sample)

	// Prepara una slice per gli N−1 punti di separazione
	ranges := make([]int, numReducers-1)

	// Cut Point equidistanti dentro al sample ordinato
	for i := 0; i < numReducers-1; i++ {
		ranges[i] = sample[i*(len(sample)/numReducers)]
	}

	fmt.Printf("ranges= %v\n", ranges)
	return ranges
}

// Assegna i chunk ai mapper e coordina le chiamate RPC
func (m *Master) ExecuteMapPhase(chunks [][]int, reducerRanges map[string][2]int) {
	var wg sync.WaitGroup        // Sincronizza tutte le goroutine
	mappers, _ := m.getMappers() // Recupera i mapper disponibili

	for i, chunk := range chunks {
		wg.Add(1)
		go func(mapper utils.WorkerConfig, chunk []int) {
			defer wg.Done()
			// Connessione RPC al mapper
			client, err := rpc.Dial("tcp", mapper.Address)
			if err != nil {
				log.Fatalf("Errore di connessione al mapper %s: %v", mapper.Address, err)
			}
			defer client.Close()
			// Costruzione della richiesta Map
			req := utils.MapRequest{Chunk: chunk, ReducerRanges: reducerRanges}
			reply := utils.MapReply{}
			// Invoca il metodo remoto MapTask sul mapper
			err = client.Call("Worker.MapTask", req, &reply)
			if err != nil {
				log.Fatalf("Errore nell'esecuzione del MapTask: %v", err)
			}
			fmt.Printf("Mapper %s ha completato l'elaborazione del chunk %v\n", mapper.Address, chunk)
		}(mappers[i%len(mappers)], chunk) // Distribuisce i chunk in round-robin tra i mapper
	}
	wg.Wait() // Attende che tutte le goroutine abbiano completato
	fmt.Println("Fase di Map completata.")
}

// ========================================================================================
// Combinazione output
// ========================================================================================

// Combina i file di output dei reducer
func (m *Master) CombineOutputFiles() {
	// Crea o sovrascrive file per risultato finale
	outputFile := "output/final_output.txt"
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Errore nella creazione del file di output: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	reducers, _ := m.getReducers()

	// Unisce i file in ordine
	for _, reducer := range reducers {

		tempFile := fmt.Sprintf("output/temp_%s.txt", strings.ReplaceAll(reducer.Address, ":", "_"))
		fmt.Printf("Unisco il file temporaneo: %s\n", tempFile)

		if _, err := os.Stat(tempFile); os.IsNotExist(err) {
			log.Printf("Errore: il file %s non esiste.\n", tempFile)
			continue
		}

		content, err := os.ReadFile(tempFile)
		if err != nil {
			log.Printf("Errore nella lettura del file %s: %v", tempFile, err)
			continue
		}

		writer.Write(content)
		writer.WriteString("\n")
	}

	writer.Flush()
	fmt.Printf("Output finale scritto in: %s\n", outputFile)
}

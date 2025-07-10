package main

import (
	"log"
	"net/rpc"
	"os/exec"
	"time"
	"sdcc-mapreduce/utils"
)

func main() {
	log.Println("[STANDBY] Avvio controller tra 10 secondi...")
	time.Sleep(10 * time.Second)

	failCount := 0

	for {
		// Se completed.json esiste, esci subito
		if utils.CompletionFlagExists() {
			log.Println("[STANDBY] Computazione completata. Nessun recovery necessario.")
			break
		}

		time.Sleep(5 * time.Second)

		client, err := rpc.Dial("tcp", "master:9000")
		if err != nil {
			failCount++
			log.Printf("[STANDBY] Tentativo fallito (%d/3)", failCount)

			// Al primo tentativo fallito, aspetta 3s e ricontrolla se il master ha terminato bene
			if failCount == 1 {
				log.Println("[STANDBY] Attendo 3 secondi per verifica completamento...")
				time.Sleep(5 * time.Second)
				if utils.CompletionFlagExists() {
					log.Println("[STANDBY] Computazione completata rilevata post-exit. Nessun recovery necessario.")
					break
				}
			}

			// Dopo 3 tentativi → recovery
			if failCount >= 3 {
				log.Println("[STANDBY] Master non risponde da 3 cicli. Avvio recovery...")
				restartMaster()
				failCount = 0
			}
			continue
		}

		// Se il master risponde → reset
		failCount = 0
		client.Close()
		log.Println("[STANDBY] Master attivo")
	}
}


func restartMaster() {
	log.Println("[STANDBY] Riavvio master...")

	cmd := exec.Command("docker", "container", "start", "master")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[STANDBY] Errore riavvio: %v\nOutput: %s", err, string(out))
	} else {
		log.Println("[STANDBY] Master riavviato con successo.")
	}
}

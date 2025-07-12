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
		// Se completed.json esiste → arresta tutto
		if utils.CompletionFlagExists() {
			log.Println("[STANDBY] Computazione completata. Arresto sistema...")
			shutdownAll()
			break
		}

		time.Sleep(7 * time.Second)

		// 🔌 Prova a connetterti al master via RPC
		client, err := rpc.Dial("tcp", "master:9000")
		if err != nil {
			failCount++
			log.Printf("[STANDBY] Tentativo fallito (%d/3)", failCount)

			// Primo tentativo fallito → attesa e controllo completamento
			if failCount == 1 {
				exec.Command("sync").Run()
				log.Println("[STANDBY] Attendo 7 secondi per verifica completamento...")
				time.Sleep(7 * time.Second)
				if utils.CompletionFlagExists() {
					log.Println("[STANDBY] Computazione completata rilevata post-exit. Arresto sistema...")
					shutdownAll()
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

func shutdownAll() {
	log.Println("[STANDBY] Arresto di tutti i container...")
	cmd := exec.Command("docker-compose", "down")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[STANDBY] Errore arresto sistema: %v\nOutput: %s", err, string(out))
	} else {
		log.Println("[STANDBY] Sistema arrestato con successo.")
	}
}


# SDCC – Fault-Tolerant MapReduce in Go

Progetto per il corso **Sistemi Distribuiti e Cloud Computing (SDCC)**, a.a. 2024/25.

Il progetto implementa un sistema distribuito di ordinamento basato sul paradigma **MapReduce**, scritto in **Go**, con comunicazione via **RPC**, e organizzato in **master**, **mapper** e **reducer**.

Tutte le componenti sono containerizzate con **Docker** e orchestrate tramite **Docker Compose**.

## Come eseguire il progetto (per ora solo in locale)

### 1. Modifica `config/config.json`

Imposta i parametri del sistema:
- `numMappers`: numero di mapper da avviare
- `numReducers`: numero di reducer da avviare
- `xi`, `xf`: range dei numeri generati (es. da 1 a 50)
- `count`: quantità totale di numeri casuali da generare
- `workers`: lista dei worker, ognuno con:
  - `role`: `"mapper"` o `"reducer"`
  - `address`: devono combaciare con i nomi host definiti in `docker-compose.yml`!

### 3. Avvia il sistema completo

```bash
./script/run.sh
```

### 4. Verifica output

- I reducer scrivono output parziale in `output/temp_<address>.txt`
- Il master unisce tutto in `output/final_output.txt`

## Strategia partizionamento

- Per bilanciare il carico tra i reducer il master esegue un sampling del 10% del dataset (per evitare di gestire troppi dati e annullare i benefici del map/reduce)
- Il sample viene ordinato e si estraggono N-1 cut point equidistanti nel sample per generare N intervalli

## Requisiti

- Docker & Docker Compose installati (es. Docker Desktop)
- Go ≥ 1.20 solo per sviluppo, non per esecuzione

## Autore

Mariani Dennis
Università di Roma Tor Vergata
Corso Sistemi Distribuiti e Cloud – A.A. 2024/25

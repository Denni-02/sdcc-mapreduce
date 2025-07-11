# SDCC – Fault-Tolerant MapReduce in Go

Progetto per il corso **Sistemi Distribuiti e Cloud Computing (SDCC)**, a.a. 2024/25.

Il progetto implementa un sistema distribuito di ordinamento basato sul paradigma **MapReduce**, scritto in **Go**, con comunicazione via **RPC**, organizzato in **master**, **mapper** e **reducer**, e con salvataggio dello stato mediante S3.

Tutte le componenti sono containerizzate con **Docker** e orchestrate tramite **Docker Compose**. Il sistema è eseguibile sia localmente, sia in cloud (EC2 AWS).

## Requisiti

### Per l'esecuzione in locale:
- Docker & Docker Compose installati
- Go ≥ 1.20 (solo per modificare o ricompilare il codice)

### Per eseguire su EC2 (Learner Lab AWS):
- Accesso a un **AWS Learner Lab**
- Una **istanza EC2** attiva nel Lab
- Docker e Docker Compose preinstallati nel Lab
- Una **chiave `.pem`** per la connessione SSH all’istanza



## Come eseguire il progetto su un'istanza EC2

### 1. Avviare Learner Lab AWS
- Entra nel **Learner Lab**
- Avvia l'istanza **EC2**
- Copia
  *  Credenziali temporanee (Access Key, Secret Key, Session Token)
  *  IP pubblico dell'istanza

### 2. Connettersi all’istanza EC2

Data la chiave `sdcc-key.pem` salvata in `~/.ssh` con permessi corretti, si si connette nel seguente modo:

```bash
ssh -i ~/.ssh/sdcc-key.pem ec2-user@<IP_EC2>
```

### 3. Entra nella cartella del progetto

Se è già presente:

```bash
cd sdcc-mapreduce
git pull
```

Altrimenti:

```bash
git clone https://github.com/Denni-02/sdcc-mapreduce.git
```

### 4. Impostare credenziali AWS

Esegui lo script helper che crea lo scheletro del file .env:

```bash
./script/init_env.sh
```

Poi apri .env e incolla le credenziali temporanee dal Lab e il nome del bucket S3:

```bash
ENABLE_S3=true
S3_BUCKET=sdcc-mapreduce-recovery
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
AWS_SESSION_TOKEN=...
```

### 5. Modifica `config/config.json`

Imposta i parametri del sistema:
- `numMappers`: numero di mapper da avviare di default
- `numReducers`: numero di reducer da avviare di default
- `xi`, `xf`: range dei numeri generati (es. da 1 a 50)
- `count`: quantità totale di numeri casuali da generare


### 6. Avvia il sistema completo

Esegui il wrapper script run.sh con eventuali argomenti per specificare il numero di mapper e reducer:

```bash
./script/run_EC2.sh numMappers numReducer
```

### 7. Verifica output

- I reducer scrivono output parziale in `output/temp_<address>.txt`
- Il master unisce tutto in `output/final_output.txt`

## Strategia partizionamento

- Per bilanciare il carico tra i reducer il master esegue un sampling del 10% del dataset (per evitare di gestire troppi dati e annullare i benefici del map/reduce)
- Il sample viene ordinato e si estraggono N-1 cut point equidistanti nel sample per generare N intervalli

## Autore

Mariani Dennis
Università di Roma Tor Vergata
Corso Sistemi Distribuiti e Cloud – A.A. 2024/25

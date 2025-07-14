# SDCC – Fault-Tolerant MapReduce in Go

Progetto per il corso **Sistemi Distribuiti e Cloud Computing (SDCC)**, a.a. 2024/25.

Il progetto implementa un sistema distribuito di ordinamento basato sul paradigma **MapReduce**, scritto in **Go**, con comunicazione via **RPC**, organizzato in **master**, **mapper** e **reducer**, e con salvataggio dello stato mediante S3.

Tutte le componenti sono containerizzate con **Docker** e orchestrate tramite **Docker Compose**. Il sistema è eseguibile sia localmente, sia in cloud (EC2 AWS).

--- 

## Requisiti

### Per l'esecuzione in locale:
- Docker & Docker Compose installati
- Go ≥ 1.20 (solo per modificare o ricompilare il codice)

### Per eseguire su EC2 (Learner Lab AWS):
- Accesso a un **AWS Learner Lab**
- Una **istanza EC2** attiva nel Lab
- Docker e Docker Compose installati nell'istanza
- Una **chiave `.pem`** per la connessione SSH all’istanza

---

## Come eseguire il progetto su un'istanza EC2

### 1. Avviare Learner Lab AWS
- Entra nel **Learner Lab**
- Avvia l'istanza **EC2**
- Copia
  *  Credenziali temporanee (Access Key, Secret Key, Session Token)
  *  IP pubblico dell'istanza

### 2. Connettersi all’istanza EC2

Data la chiave `sdcc-key.pem` salvata in `~/.ssh` con permessi corretti, ci si connette nel seguente modo:

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
cd sdcc-mapreduce
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

Se necessario rendilo eseguibile nel seguente modo:
```bash
chmod +x script/init_env.sh
```

### 5. Modifica `config/config.json`

Imposta i parametri del sistema:
- `xi`, `xf`: range dei numeri generati (es. da 1 a 50)
- `count`: quantità totale di numeri casuali da generare


### 6. Avvia il sistema completo

Esegui lo script run_EC2.sh con eventuali argomenti per specificare il numero di mapper e reducer (altrimenti parte con un numero di default definito in config.json):

```bash
./script/run_EC2.sh numMappers numReducer
```
Se necessario rendilo eseguibile nel seguente modo:
```bash
chmod +x script/run_EC2.sh
```

### 7. Verifica output

- I reducer scrivono output parziale in `output/temp_<address>.txt`
- Il master unisce tutto in `output/final_output.txt`

E' possibile utilizzare gli script view_output.sh e view_master_log.sh:

Se necessario rendilo eseguibile nel seguente modo:
```bash
chmod +x script/view_output.sh
./script/view_output.sh
chmod +x script/view_master_log.sh
./script/view_master_log.sh
```
---

## Strategia partizionamento

- Per bilanciare il carico tra i reducer il master esegue un sampling del 10% del dataset (per evitare di gestire troppi dati e annullare i benefici del map/reduce)
- Il sample viene ordinato e si estraggono N-1 cut point equidistanti nel sample per generare N intervalli

--- 

## Guida creazione EC2 (se necessario)

È possibile eseguire l’intero sistema in cloud su Amazon EC2, sfruttando gli strumenti offerti dal Learner Lab di AWS Academy.

1. Accedere a AWS Academy (https://www.awsacademy.com/vforcesite/LMS_Login)
2. Entra nel corso e clicca su Start Lab
3. Clicca su AWS per aprire la console
4. Crea EC2 andando in Servizi → EC2 → Launch Instance, per esempio io ho utilizzato i seguenti parametri:
  * Name: sdcc-master
  * AMI: Amazon Linux 2023 (64-bit x86)
  * Tipo: t2.micro
  * Key Pair: sdcc-key.pem (scaricala e conservala)
  * Security Group 
    - Regola 1: SSH (porta 22) da Anywhere 
    - Regola 2: HTTP (porta 80) da Anywhere
  * Avvia la macchina e copia l’indirizzo Public IPv4
6. Sposta chiave nella directory .ssh
5. Collegati via SSH
```bash
chmod 400 ~/.ssh/sdcc-key.pem
ssh -i ~/.ssh/sdcc-key.pem ec2-user@<IP_PUBBLICO>
```
6. Installa Docker e Docker Compose
```bash
sudo yum update -y
sudo yum install -y docker git
sudo service docker start
sudo usermod -a -G docker ec2-user
newgrp docker
sudo curl -L "https://github.com/docker/compose/releases/download/v2.27.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
docker-compose version
```

## Guida creazione bucket S3 (se necessario)
1. Vai su: Services → S3 → Create bucket
2. Per esempio inserisci i seguenti valori:
   * Bucket name: sdcc-mapreduce-recovery
   * Region: lasciarla di default (es: us-east-1)
   * ACLs enabled 
   * Block all public access: deseleziona tutte le opzioni
   * Conferma cliccando su Create bucket
   * Vai in Permissions → Bucket policy e inserisci

```bash
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowPublicAll",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:*",
      "Resource": "arn:aws:s3:::sdcc-mapreduce-recovery/*"
    }
  ]
}

```

--- 

# Autore

Mariani Dennis
Università di Roma Tor Vergata
Corso Sistemi Distribuiti e Cloud – A.A. 2024/25

FROM golang:1.23.3-alpine

# Installa Docker CLI per usare `docker compose` nel container
RUN apk add --no-cache docker-cli

WORKDIR /app

# Copia l'intero progetto (serve per risolvere le importazioni locali come sdcc-mapreduce/utils)
COPY . .

WORKDIR /app/standby

# Costruzione del binario standby
RUN go build -o standby main.go

# Esegue il binario
CMD ["./standby"]

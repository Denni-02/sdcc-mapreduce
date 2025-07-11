FROM golang:1.23.3-alpine

# Installa Docker CLI per usare `docker compose` nel container
RUN apk add --no-cache docker-cli docker-compose

WORKDIR /app

# Copia l'intero progetto
COPY . .

# Costruzione del binario chiamato standby_bin
RUN go build -o standby_bin ./standby/main.go

# Avvio del binario
CMD ["./standby_bin"]

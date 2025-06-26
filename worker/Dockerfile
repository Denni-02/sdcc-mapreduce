FROM golang:1.23.3-alpine

WORKDIR /app

COPY . .

RUN go build -o worker_bin ./worker          

CMD ["sh", "-c", "./worker_bin --address=${HOSTNAME}:900$ID"]



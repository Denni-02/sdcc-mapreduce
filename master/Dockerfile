FROM golang:1.23.3-alpine

WORKDIR /app

COPY . .

RUN go build -o master_bin ./master

CMD ["./master_bin"]

FROM debian:bullseye-slim

# Imposta working directory
WORKDIR /app

# Installa dipendenze
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    zip \
    bash \
    less \
    groff \
    awscli \
    ca-certificates \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copia tutto il codice
COPY . .

# Installa Go (manuale perché non incluso)
RUN curl -LO https://golang.org/dl/go1.21.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz && \
    rm go1.21.0.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:$PATH"

# Compila il codice
RUN go build -o master_bin ./master

# Entry point
CMD ["./master_bin"]

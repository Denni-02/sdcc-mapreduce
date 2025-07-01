#!/bin/bash

# Numero di mapper e reducer da usare (default: 4)
NUM_MAPPERS=${1:-4}
NUM_REDUCERS=${2:-4}

echo "🧹 Pulizia container precedenti..."
docker-compose down --volumes --remove-orphans

echo "🔨 Build immagini Docker..."
docker-compose build

echo "🚀 Avvio con $NUM_MAPPERS mapper e $NUM_REDUCERS reducer..."
docker-compose up -d --scale mapper=$NUM_MAPPERS --scale reducer=$NUM_REDUCERS

echo "✅ Sistema MapReduce avviato."

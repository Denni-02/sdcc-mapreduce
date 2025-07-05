#!/bin/bash

set -a
source .env
set +a

# Numero di mapper e reducer passati da linea di comando, default = 4
NUM_MAPPERS=${1:-4}
NUM_REDUCERS=${2:-4}
CONFIG_PATH="./config/config.json"

echo "Pulizia container precedenti..."
docker-compose down --volumes --remove-orphans

echo "Pulizia log..."
rm -f ./log/log_master/*.log
rm -f ./log/log_worker/*.log

echo "Aggiornamento config.json con $NUM_MAPPERS mapper e $NUM_REDUCERS reducer..."

# Inserisce dinamicamente numMappers e numReducers nel config.json
jq ".settings.numMappers = $NUM_MAPPERS | .settings.numReducers = $NUM_REDUCERS" \
  "$CONFIG_PATH" > "$CONFIG_PATH.tmp" && mv "$CONFIG_PATH.tmp" "$CONFIG_PATH"

echo "Build immagini Docker..."
docker-compose build

echo "Avvio con $NUM_MAPPERS mapper e $NUM_REDUCERS reducer..."
docker-compose up --scale mapper=$NUM_MAPPERS --scale reducer=$NUM_REDUCERS

echo "Sistema avviato."

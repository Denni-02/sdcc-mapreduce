#!/bin/bash

echo "Count,Time(s)" > benchmark_results.csv

NUM_MAPPERS=4
NUM_REDUCERS=4

for COUNT in 100 1000 5000 10000; do
  echo "Esecuzione con count=$COUNT..."

  # Aggiorna solo il campo "count" nel config.json
  jq ".settings.count = $COUNT" ../config/config.json > tmp.json && mv tmp.json ../config/config.json

  # Pulizia output
  rm -f ../output/*

  # Tempo di inizio
  START=$(date +%s)

  # Esegui con timeout e passa i parametri corretti
  timeout 300 ../script/run.sh $NUM_MAPPERS $NUM_REDUCERS > /dev/null 2>&1
  STATUS=$?

  # Tempo di fine
  END=$(date +%s)
  DIFF=$((END - START))

  if [ $STATUS -eq 124 ]; then
    echo "Timeout su count=$COUNT"
    echo "$COUNT,TIMEOUT" >> benchmark_results.csv
  else
    echo "$COUNT,$DIFF" >> benchmark_results.csv
    echo "count=$COUNT completato in ${DIFF}s"
  fi

  docker-compose down --remove-orphans > /dev/null 2>&1
done

echo "Benchmark completato! Risultati salvati in benchmark_results.csv"

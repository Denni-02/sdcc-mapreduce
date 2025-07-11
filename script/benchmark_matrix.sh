#!/bin/bash

sudo chown -R $USER:$USER ../log ../output
sudo chmod -R u+w ../log ../output

# Combinazioni da testare
MAPPERS_LIST=(2 4 8)
REDUCERS_LIST=(2 4 8)
COUNTS=(100 1000 5000 10000 20000)

CONFIG_PATH="../config/config.json"
RESULT_FILE="benchmark_results.csv"

# Inizializza CSV
echo "Mappers,Reducers,Count,Time(s)" > "$RESULT_FILE"

for M in "${MAPPERS_LIST[@]}"; do
  for R in "${REDUCERS_LIST[@]}"; do
    for COUNT in "${COUNTS[@]}"; do

      echo "Mappers=$M Reducers=$R Count=$COUNT"

      # Aggiorna il campo count in config.json
      jq ".settings.count = $COUNT" "$CONFIG_PATH" > tmp.json && mv tmp.json "$CONFIG_PATH"

      # Pulizia output
      sudo rm -f ../output/*

      # Tempo di inizio
      START=$(date +%s)

      # Esegui run.sh con timeout e mappers/reducers dinamici
      timeout 300 ../script/run.sh "$M" "$R" > /dev/null 2>&1
      STATUS=$?

      # Tempo di fine
      END=$(date +%s)
      DIFF=$((END - START))

      # Scrivi su CSV
      if [ $STATUS -eq 124 ]; then
        echo "$M,$R,$COUNT,TIMEOUT" >> "$RESULT_FILE"
        echo "Timeout"
      else
        echo "$M,$R,$COUNT,$DIFF" >> "$RESULT_FILE"
        echo "Completato in ${DIFF}s"
      fi

      # Stop containers
      docker-compose down --remove-orphans > /dev/null 2>&1
    done
  done
done

echo "Benchmark completato: risultati in $RESULT_FILE"

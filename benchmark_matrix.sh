#!/bin/bash

# Fix permessi
sudo chown -R $USER:$USER ./log ./output ./state
sudo chmod -R u+w ./log ./output ./state

# Combinazioni da testare
MAPPERS_LIST=(2 4 8)
REDUCERS_LIST=(2 4 8)
COUNTS=(100 1000 5000 10000 20000)

CONFIG_PATH="./config/config.json"
RESULT_FILE="benchmark_results.csv"

# Inizializza CSV
echo "Mappers,Reducers,Count,Time(s)" > "$RESULT_FILE"

for M in "${MAPPERS_LIST[@]}"; do
  for R in "${REDUCERS_LIST[@]}"; do
    for COUNT in "${COUNTS[@]}"; do

      echo "Mappers=$M Reducers=$R Count=$COUNT"

      jq ".settings.count = $COUNT" "$CONFIG_PATH" > tmp.json && mv tmp.json "$CONFIG_PATH"
      sudo rm -f ./output/*
      sudo rm -f ./state/completed.json

      START=$(date +%s)

      # Avvia run.sh in background
      ./script/run.sh "$M" "$R" &
      RUN_PID=$!

      MAX_TIME=300  # timeout
      SECONDS=0

      # Attendi che appaia completed.json o scada il tempo
      while [ ! -f ./state/completed.json ]; do
        sleep 1
        ((SECONDS++))
        if [ $SECONDS -ge $MAX_TIME ]; then
          echo "Timeout: uccido run.sh (PID $RUN_PID)"
          kill -INT $RUN_PID 2>/dev/null
          STATUS=124
          break
        fi
      done

      # Se trovato il file, aspetta che run.sh termini da solo
      if [ -f ./state/completed.json ]; then
        echo "completed.json trovato, aspetto che run.sh termini da solo..."

        for i in {1..15}; do
          if ! kill -0 $RUN_PID 2>/dev/null; then
            echo "run.sh ha terminato spontaneamente."
            STATUS=0
            break
          fi
          sleep 1
        done

        # Se dopo 15 secondi è ancora attivo, lo chiudo manualmente
        if kill -0 $RUN_PID 2>/dev/null; then
          echo "⌛ run.sh ancora attivo dopo 15s, termino docker-compose forzatamente"
          pkill -f "docker-compose up"
          wait $RUN_PID
          STATUS=0
        fi
      fi


      END=$(date +%s)
      DIFF=$((END - START))

      if [ "$STATUS" == "124" ]; then
        echo "$M,$R,$COUNT,TIMEOUT" >> "$RESULT_FILE"
        echo "Timeout"
      else
        echo "$M,$R,$COUNT,$DIFF" >> "$RESULT_FILE"
        echo "Completato in ${DIFF}s"
      fi

      docker-compose down --remove-orphans > /dev/null 2>&1

    done
  done
done

echo "Benchmark completato: risultati in $RESULT_FILE"

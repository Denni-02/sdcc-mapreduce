#!/bin/bash

# ermessi iniziali
sudo chown -R $USER:$USER ./log ./output ./state
sudo chmod -R u+w ./log ./output ./state

# Parametri di benchmark
MAPPERS_LIST=(2 4 8)
REDUCERS_LIST=(2 4 8)
COUNTS=(100 1000 5000 10000 20000)

CONFIG_PATH="./config/config.json"
RESULT_FILE="benchmark_results2.csv"

# CSV header
echo "Mappers,Reducers,Count,Time(s)" > "$RESULT_FILE"

for M in "${MAPPERS_LIST[@]}"; do
  for R in "${REDUCERS_LIST[@]}"; do
    for COUNT in "${COUNTS[@]}"; do

      echo ""
      echo "Avvio benchmark: Mappers=$M Reducers=$R Count=$COUNT"

      # Aggiorna config.json
      jq ".settings.numMappers = $M | .settings.numReducers = $R | .settings.count = $COUNT" "$CONFIG_PATH" > tmp.json \
        && mv tmp.json "$CONFIG_PATH"

      # Pulizia file locali
      sudo rm -f ./output/*
      sudo rm -f ./state/completed.json

      START=$(date +%s)

      # Avvia docker-compose (senza rebuild)
      docker-compose up --scale mapper=$M --scale reducer=$R &
      COMPOSE_PID=$!

      # Attendi completion o timeout
      TIMEOUT=300
      SECONDS=0
      while [ ! -f ./state/completed.json ]; do
        sleep 1
        ((SECONDS++))
        if [ $SECONDS -ge $TIMEOUT ]; then
          echo "Timeout raggiunto, termino docker-compose"
          kill -INT $COMPOSE_PID 2>/dev/null
          pkill -f "docker-compose up" 2>/dev/null
          STATUS=124
          break
        fi
      done

      # Attendi che docker-compose si chiuda da solo (se non già terminato)
      if [ -f ./state/completed.json ]; then
        echo "Computazione completata, attendo lo stop dei container..."
        for i in {1..15}; do
          if ! kill -0 $COMPOSE_PID 2>/dev/null; then
            echo "docker-compose terminato"
            STATUS=0
            break
          fi
          sleep 1
        done

        if kill -0 $COMPOSE_PID 2>/dev/null; then
          echo "docker-compose ancora attivo, forzo terminazione"
          kill -INT $COMPOSE_PID 2>/dev/null
          pkill -f "docker-compose up" 2>/dev/null
          STATUS=0
        fi
        wait $COMPOSE_PID 2>/dev/null
      fi

      END=$(date +%s)
      DIFF=$((END - START))

      # Scrivi su CSV
      if [ "$STATUS" == "124" ]; then
        echo "$M,$R,$COUNT,TIMEOUT" >> "$RESULT_FILE"
      else
        echo "$M,$R,$COUNT,$DIFF" >> "$RESULT_FILE"
      fi

      echo "Completato in ${DIFF}s"

      # Stop container per sicurezza
      docker-compose down --remove-orphans > /dev/null 2>&1

    done
  done
done

echo ""
echo "Benchmark completato. Risultati in $RESULT_FILE"

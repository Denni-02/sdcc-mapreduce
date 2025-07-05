#!/bin/bash

echo "🛠️ [FAULT INJECTION] Cerco un reducer da uccidere..."
sleep 2

# Trova un container reducer a caso
TARGET=$(docker ps --format '{{.Names}}' | grep reducer | shuf -n 1)

if [ -z "$TARGET" ]; then
  echo "Nessun reducer attivo trovato."
  exit 1
fi

echo "🔪 Uccido il reducer: $TARGET"
docker kill "$TARGET"

echo "Reducer $TARGET ucciso con successo."

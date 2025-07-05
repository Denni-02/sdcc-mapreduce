#!/bin/bash

echo "🛠️ [FAULT INJECTION] Cerco un mapper da uccidere..."
sleep 2

# Trova un container mapper a caso
TARGET=$(docker ps --format '{{.Names}}' | grep mapper | shuf -n 1)

if [ -z "$TARGET" ]; then
  echo "Nessun mapper attivo trovato."
  exit 1
fi

echo "Uccido il mapper: $TARGET"
docker kill "$TARGET"

echo "Mapper $TARGET ucciso con successo."

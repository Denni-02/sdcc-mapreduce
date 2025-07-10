#!/bin/bash

echo "Pulizia cache Docker..."

# Cancella le immagini dangling (intermedie non taggate)
docker image prune -f

# Cancella build cache (serve Docker 20.10+ con BuildKit attivo)
docker builder prune -a -f

echo "Cache Docker pulita."


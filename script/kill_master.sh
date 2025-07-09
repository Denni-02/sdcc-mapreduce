#!/bin/bash

echo "KILL del master..."
docker kill master
sleep 3

echo "Riavvio del master..."
docker compose up -d master
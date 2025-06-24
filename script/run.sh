#!/bin/bash

echo "Cleaning up old containers..."
docker-compose down --volumes --remove-orphans

echo "Building fresh containers..."
docker-compose build --no-cache

echo "Starting MapReduce system..."
docker-compose up

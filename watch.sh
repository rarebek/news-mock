#!/bin/bash

# Define the directory to watch
WATCHED_DIR="."

# Define the Docker Compose command to run when changes are detected
COMPOSE_CMD="docker-compose up -d --build"

# Run the initial Docker Compose command
$COMPOSE_CMD

# Use inotifywait to monitor the directory for changes
inotifywait -m -r -e modify,create,delete,move $WATCHED_DIR |
while read -r directory events filename; do
    echo "Change detected in $directory$filename. Rebuilding and restarting services..."
    $COMPOSE_CMD
done

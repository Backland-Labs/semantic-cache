#!/bin/sh

# start the ollama server in background
/bin/ollama serve &

#record PID
PID=$!

# wait for ollama server to start
sleep 5

echo "⛴️ Pulling mode...."
ollama pull nomic-embed-text
echo "🍻 Done"

# wait for ollama server to finish
wait $PID
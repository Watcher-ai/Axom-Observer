#!/bin/bash

echo "🚀 Starting AI Traffic Test"
echo "=========================="

# Start the sidecar in background
echo "📡 Starting Axom Observer (Sidecar)..."
./axom-observer --backend-url="http://localhost:8080/api/v1/webhook/signals" --customer-id="test-customer-456" --agent-id="test-agent-456" --http-port="8888" > sidecar.log 2>&1 &
SIDECAR_PID=$!

# Wait for sidecar to start
echo "⏳ Waiting for sidecar to start..."
sleep 5

# Test 1: HTTP traffic through sidecar
echo "🧪 Test 1: HTTP Chat Completions"
curl -X POST http://localhost:8888/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-test123" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "Hello, test message 1"}], "max_tokens": 50}' \
  -w "\nStatus: %{http_code}\n" \
  -s

echo ""
echo "🧪 Test 2: HTTP Completions"
curl -X POST http://localhost:8888/v1/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-test123" \
  -d '{"model": "gpt-3.5-turbo", "prompt": "Hello world", "max_tokens": 50}' \
  -w "\nStatus: %{http_code}\n" \
  -s

echo ""
echo "🧪 Test 3: HTTP Embeddings"
curl -X POST http://localhost:8888/v1/embeddings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-test123" \
  -d '{"model": "text-embedding-ada-002", "input": "Hello world"}' \
  -w "\nStatus: %{http_code}\n" \
  -s

echo ""
echo "🧪 Test 4: HTTP Models List"
curl -X GET http://localhost:8888/v1/models \
  -H "Authorization: Bearer sk-test123" \
  -w "\nStatus: %{http_code}\n" \
  -s

# Wait a moment for signals to be processed
echo ""
echo "⏳ Waiting for signals to be processed..."
sleep 3

# Check sidecar logs
echo ""
echo "📊 Sidecar Logs:"
echo "================"
tail -10 sidecar.log

# Check portal backend logs
echo ""
echo "📊 Portal Backend Logs:"
echo "======================"
cd /Users/vishesh/Documents/AgentOp/portal
docker-compose logs backend | grep -E "(webhook|signal|processed)" | tail -5

# Test the portal frontend to see if signals are displayed
echo ""
echo "🌐 Testing Portal Frontend..."
echo "============================="
echo "Open http://localhost:3000 in your browser to see the signals in the dashboard"

# Cleanup
echo ""
echo "🧹 Cleaning up..."
kill $SIDECAR_PID 2>/dev/null

echo ""
echo "✅ Test completed!"
echo "Check the portal dashboard at http://localhost:3000 to see the processed signals"


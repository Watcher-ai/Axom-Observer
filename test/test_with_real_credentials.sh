#!/bin/bash

# Simple test script - Observer + Mock AI Agent
set -e

echo "🚀 Starting Axom Observer + Mock AI Agent"
echo "========================================="

# Your agent credentials
export CUSTOMER_ID="t1"
export AGENT_ID="76f0ca92-ecbd-4697-a421-36af53fa891f"
export CLIENT_ID="8d9bd6fb-ef0e-4ee8-878b-ce21a6ade460"
export CLIENT_SECRET="cXq5gzMOcizzG9p9XiAjJw617MldfbIw"
export AGENT_SECRET="hmR2hg8z0ZXuXI0eUTt4WOmDVqbzNLiy"
export BACKEND_URL="http://localhost:8080/api/v1/webhook/signals"
export LOG_LEVEL="debug"
export LOG_ALL_TRAFFIC="true"

echo "🔑 Agent: $CUSTOMER_ID | $AGENT_ID"
echo "🌐 Webhook: $BACKEND_URL"
echo ""

# Build and start observer
echo "🔨 Building observer..."
cd ..
docker build -t axom-observer . > /dev/null 2>&1
cd test

echo "🚀 Starting observer..."
docker run -d \
    --name axom-observer \
    -p 8888:8888 \
    -p 8443:8443 \
    -e CUSTOMER_ID="$CUSTOMER_ID" \
    -e AGENT_ID="$AGENT_ID" \
    -e CLIENT_ID="$CLIENT_ID" \
    -e CLIENT_SECRET="$CLIENT_SECRET" \
    -e BACKEND_URL="$BACKEND_URL" \
    -e LOG_LEVEL="$LOG_LEVEL" \
    -e LOG_ALL_TRAFFIC="$LOG_ALL_TRAFFIC" \
    axom-observer

sleep 3
echo "✅ Observer running on ports 8888 (HTTP) and 8443 (HTTPS)"

# Setup Python environment
if [ ! -d "venv" ]; then
    python3 -m venv venv
    source venv/bin/activate
    pip install requests > /dev/null 2>&1
else
    source venv/bin/activate
fi

echo "🤖 Starting mock AI agent..."
echo ""

# Run mock AI agent
python3 mock_ai_agent.py --observer-url http://localhost:8888 --test all

echo ""
echo "✅ Test completed! Check your portal for received signals."
echo ""
echo "📋 Observer logs:"
docker logs axom-observer --tail 10

echo ""
echo "🛑 To stop: docker stop axom-observer && docker rm axom-observer"

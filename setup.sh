#!/bin/bash

# Axom Observer Setup Script
# This script helps you configure the Axom Observer with your agent details

set -e

echo "🚀 Axom Observer Setup"
echo "======================"
echo ""
echo "This script will help you configure the Axom Observer with your agent details."
echo "You can find these details in your Axom Portal under 'Agent Details'."
echo ""

# Check if .env file already exists
if [ -f ".env" ]; then
    echo "⚠️  .env file already exists!"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled."
        exit 0
    fi
fi

# Function to prompt for required input
prompt_input() {
    local var_name=$1
    local prompt_text=$2
    local is_secret=${3:-false}
    
    while true; do
        if [ "$is_secret" = true ]; then
            read -s -p "$prompt_text: " value
            echo
        else
            read -p "$prompt_text: " value
        fi
        
        if [ -n "$value" ]; then
            echo "$var_name=$value" >> .env
            break
        else
            echo "❌ This field is required. Please try again."
        fi
    done
}

# Create .env file
echo "# Axom Observer Configuration" > .env
echo "# Generated on $(date)" >> .env
echo "" >> .env

echo "📝 Please enter your agent details:"
echo ""

# Prompt for required fields
prompt_input "CUSTOMER_ID" "Agent Name (from your portal)"
prompt_input "AGENT_ID" "Agent ID (from your portal)"
prompt_input "CLIENT_ID" "Client ID (from your portal)"
prompt_input "CLIENT_SECRET" "Client Secret (from your portal)" true

echo ""
echo "🔧 Optional configuration:"
echo ""

# Prompt for optional fields
read -p "Backend URL (default: https://api.axom.ai/ingest): " backend_url
if [ -n "$backend_url" ]; then
    echo "BACKEND_URL=$backend_url" >> .env
else
    echo "BACKEND_URL=https://api.axom.ai/ingest" >> .env
fi

read -p "Log Level (default: info): " log_level
if [ -n "$log_level" ]; then
    echo "LOG_LEVEL=$log_level" >> .env
else
    echo "LOG_LEVEL=info" >> .env
fi

read -p "Log All Traffic (default: true): " log_all_traffic
if [ -n "$log_all_traffic" ]; then
    echo "LOG_ALL_TRAFFIC=$log_all_traffic" >> .env
else
    echo "LOG_ALL_TRAFFIC=true" >> .env
fi

read -p "Main AI Container Name (default: my-ai-app): " main_container
if [ -n "$main_container" ]; then
    echo "MAIN_AI_CONTAINER_NAME=$main_container" >> .env
else
    echo "MAIN_AI_CONTAINER_NAME=my-ai-app" >> .env
fi

echo ""
echo "✅ Configuration saved to .env file!"
echo ""
echo "📋 Your configuration:"
echo "======================"
cat .env | grep -v "CLIENT_SECRET" | sed 's/CLIENT_SECRET=.*/CLIENT_SECRET=***HIDDEN***/'
echo ""
echo "🚀 You can now run: docker-compose up"
echo ""
echo "💡 To modify your configuration later, edit the .env file or run this script again."

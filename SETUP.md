# Axom Observer Setup Guide

This guide will help you configure the Axom Observer with your agent details from the Axom Portal.

## Quick Start

### Option 1: Interactive Setup (Recommended)

Run the interactive setup script:

```bash
./setup.sh
```

This script will prompt you for all required information and create a `.env` file automatically.

### Option 2: Manual Configuration

1. Copy the template file:
   ```bash
   cp env.template .env
   ```

2. Edit the `.env` file with your actual values:
   ```bash
   nano .env
   ```

3. Fill in your agent details from the Axom Portal:
   - `CUSTOMER_ID` - Your Agent Name
   - `AGENT_ID` - Your Agent ID  
   - `CLIENT_ID` - Your Client ID
   - `CLIENT_SECRET` - Your Client Secret

### Option 3: Environment Variables

Set the environment variables directly:

```bash
export CUSTOMER_ID="your-agent-name"
export AGENT_ID="your-agent-id"
export CLIENT_ID="your-client-id"
export CLIENT_SECRET="your-client-secret"
export BACKEND_URL="https://api.axom.ai/ingest"
```

## Required Configuration

The following fields are **required** and must be obtained from your Axom Portal:

| Field | Description | Example |
|-------|-------------|---------|
| `CUSTOMER_ID` | Your Agent Name | `aaa` |
| `AGENT_ID` | Your Agent ID | `e75b6a8c-d7af-464b-89ca-3eeff8b36715` |
| `CLIENT_ID` | Your Client ID | `6332d145-cf73-49a1-bd26-4fbda842411c` |
| `CLIENT_SECRET` | Your Client Secret | `5ukEa4hEsL0LBhVa9QkPgTyDCQzauccp` |

## Optional Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `BACKEND_URL` | Backend URL for signals | `https://api.axom.ai/ingest` |
| `LOG_LEVEL` | Logging level | `info` |
| `LOG_ALL_TRAFFIC` | Log all traffic (debug) | `true` |
| `MAIN_AI_CONTAINER_NAME` | Main AI container name | `my-ai-app` |

## Running the Observer

Once configured, start the observer:

```bash
docker-compose up
```

The observer will:
- Validate all required configuration
- Start HTTP proxy on port 8888
- Start HTTPS proxy on port 8443
- Begin monitoring AI API traffic
- Send signals to your backend

## Verification

After starting, you should see logs like:

```
🚀 Starting Axom AI Observer
📡 Customer ID: aaa
🤖 Agent ID: e75b6a8c-d7af-464b-89ca-3eeff8b36715
🔑 Client ID: 6332d145-cf73-49a1-bd26-4fbda842411c
🔐 Client Secret: 5ukE***ccp
🌐 Backend URL: https://api.axom.ai/ingest
🔗 HTTP Port: 8888
🔒 HTTPS Port: 8443
✅ Observer started successfully
```

## Troubleshooting

### Missing Configuration Error

If you see:
```
❌ Missing required configuration!
Please provide the following environment variables:
  CUSTOMER_ID    - Your Agent Name
  AGENT_ID       - Your Agent ID
  CLIENT_ID      - Your Client ID
  CLIENT_SECRET  - Your Client Secret
```

This means one or more required fields are missing. Run `./setup.sh` to configure them.

### No Signals Being Sent

1. Verify your agent details are correct
2. Check that your AI application is routing traffic through the observer ports (8888 for HTTP, 8443 for HTTPS)
3. Ensure the backend URL is accessible
4. Check the logs for any authentication errors

## Security Notes

- Never commit your `.env` file to version control
- Keep your `CLIENT_SECRET` secure
- The observer masks sensitive information in logs
- Use environment variables in production deployments

## Support

If you encounter issues:
1. Check the logs for error messages
2. Verify your agent details in the Axom Portal
3. Ensure all required environment variables are set
4. Contact support with your agent ID and error details

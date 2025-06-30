# Axom Observer - AI Traffic Monitoring Sidecar

A production-grade MITM proxy for monitoring AI API traffic in real-time. Designed to run as a sidecar container alongside your AI applications.

## 🚀 Features

- **Production MITM Proxy**: Built with `gomitmproxy` for robust HTTPS interception
- **AI Provider Detection**: Automatically detects 20+ AI service providers
- **Real-time Monitoring**: Captures AI requests/responses with latency metrics
- **Task Detection**: Identifies and groups related AI operations
- **Containerized**: Ready for sidecar deployment in Kubernetes/Docker
- **Multi-protocol**: Supports HTTP, HTTPS, and WebSocket traffic

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Your AI App   │───▶│  Axom Observer  │───▶│  AI Providers   │
│                 │    │   (Sidecar)     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Backend API   │
                       │  (Signals)      │
                       └─────────────────┘
```

## 🐳 Quick Start with Docker

### 1. Build and Run with Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd Axom-Observer

# Build and start all services
docker-compose up -d

# Check service status
docker-compose ps
```

### 2. Test the Setup

```bash
# Run comprehensive tests
chmod +x test_containerized.sh
./test_containerized.sh

# Or test individual components
chmod +x test_real_ai_endpoints.sh
./test_real_ai_endpoints.sh
```

### 3. Access Services

- **Observer HTTP Proxy**: http://localhost:8888
- **Observer HTTPS Proxy**: http://localhost:8443
- **Demo AI App**: http://localhost:5002
- **Backend Dashboard**: http://localhost:8080

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CUSTOMER_ID` | `test-customer` | Your customer identifier |
| `AGENT_ID` | `test-agent` | Your agent identifier |
| `BACKEND_URL` | `http://localhost:8080/api/v1/signals` | Backend API endpoint |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |

### Docker Compose Configuration

```yaml
services:
  observer:
    build: .
    environment:
      - CUSTOMER_ID=your-customer-id
      - AGENT_ID=your-agent-id
      - BACKEND_URL=http://your-backend/api/v1/signals
    ports:
      - "8888:8888"   # HTTP proxy
      - "8443:8443"   # HTTPS proxy
```

## 🚀 Sidecar Deployment

### Kubernetes Deployment

```bash
# Apply the sidecar deployment
kubectl apply -f k8s/observer-sidecar.yaml

# Check deployment status
kubectl get pods -l app=ai-app-with-observer
```

### Custom Application Integration

To use the observer as a sidecar with your application:

1. **Set proxy environment variables**:
```bash
export HTTP_PROXY=http://localhost:8888
export HTTPS_PROXY=http://localhost:8443
export NO_PROXY=localhost,127.0.0.1
```

2. **Configure your AI client**:
```python
import requests

# Your AI requests will automatically go through the proxy
response = requests.post(
    "https://api.openai.com/v1/chat/completions",
    json={
        "model": "gpt-4",
        "messages": [{"role": "user", "content": "Hello"}]
    },
    proxies={
        "http": "http://localhost:8888",
        "https": "http://localhost:8443"
    }
)
```

## 🔍 Supported AI Providers

The observer automatically detects traffic from:

### LLM Providers
- OpenAI (api.openai.com)
- Anthropic (api.anthropic.com)
- Google AI (generativelanguage.googleapis.com)
- Cohere (api.cohere.ai)
- Together AI (api.together.ai)
- Groq (api.groq.com)
- Hugging Face (api-inference.huggingface.co)
- Azure OpenAI (*.openai.azure.com)

### Speech Services
- Deepgram (api.deepgram.com)
- AssemblyAI (api.assemblyai.com)
- ElevenLabs (api.elevenlabs.io)
- PlayHT (api.play.ht)
- Amazon Polly (polly.*.amazonaws.com)
- Azure TTS (*.cognitiveservices.azure.com)

### Communication Services
- Twilio (api.twilio.com)
- Plivo (api.plivo.com)
- Vonage (api.nexmo.com, api.vonage.com)
- Daily (api.daily.co)
- 100ms (api.100ms.live)

## 🧪 Testing

### Test with Real AI Endpoints

```bash
# Test with free AI endpoints
./test_real_ai_endpoints.sh

# Test with demo app
curl -X POST http://localhost:5002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 50
  }'
```

### Test with Your Own AI Applications

1. **Configure your app to use the proxy**:
```bash
export HTTP_PROXY=http://localhost:8888
export HTTPS_PROXY=http://localhost:8443
```

2. **Make AI API calls** - they'll be automatically captured

3. **Check the observer logs**:
```bash
docker logs axom-observer
```

## 🔒 Security Considerations

### Certificate Management

The observer uses `gomitmproxy`'s built-in CA certificate. For production:

1. **Trust the CA certificate** in your client applications
2. **Use custom certificates** if needed (requires `gomitmproxy` patching)
3. **Monitor certificate expiration**

### Network Security

- Run the observer in a secure network environment
- Use network policies to restrict access
- Monitor for unauthorized proxy usage

## 🛠️ Development

### Building from Source

```bash
# Build the observer
go build -o observer main.go

# Run locally
./observer --customer-id="dev" --agent-id="local"
```

### Adding New AI Providers

Edit `pkg/observer/ai_traffic_monitor.go` to add new providers:

```go
{
    Name: "New Provider",
    Domains: []string{"api.newprovider.com"},
    APIPatterns: []string{"/v1/chat", "/v1/embed"},
},
```

## 📈 Production Deployment

### Resource Requirements

- **CPU**: 100-500m (depending on traffic volume)
- **Memory**: 128-512Mi (depending on traffic volume)
- **Storage**: Minimal (logs and metrics)

### Scaling Considerations

- Deploy one observer per application pod
- Use horizontal pod autoscaling for high-traffic applications
- Monitor resource usage and adjust limits accordingly

### High Availability

- Use Kubernetes deployments with multiple replicas
- Implement proper health checks and readiness probes
- Set up monitoring and alerting

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🆘 Support

- **Issues**: Create an issue on GitHub
- **Documentation**: Check the docs folder
- **Examples**: See the demo and test directories

---

**Note**: This observer is designed for monitoring and debugging purposes. Ensure compliance with your organization's security policies and data handling requirements.

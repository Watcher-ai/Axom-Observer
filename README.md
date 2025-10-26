# Axom Observer - Plug-and-Play AI Traffic Sidecar

A production-ready MITM proxy sidecar for capturing **all HTTP and HTTPS traffic** from your AI agent (or any app) and forwarding it to your backend/webhook. No code changes required—just set a few environment variables!

---

## 🚀 Quick Start (Docker Compose)

```bash
# Clone the repository
git clone <repository-url>
cd Axom-Observer

# Build and start all services
docker-compose up -d
```

---

## 🧩 Plug-and-Play Integration with Any AI Agent

1. **Add the observer as a sidecar** (in Docker Compose or Kubernetes).
2. **Set these environment variables in your AI agent container:**
   ```bash
   export HTTP_PROXY=http://observer:8888
   export HTTPS_PROXY=http://observer:8443
   export NO_PROXY=localhost,127.0.0.1
   ```
   - Replace `observer` with the service name or `localhost` if running locally.

3. **(For HTTPS interception) Trust the observer's CA certificate:**
   - The observer generates a CA cert at startup (see `certs/` directory).
   - Add this CA to your agent's trust store:
     - **Python:**
       ```python
       import certifi
       # Add the observer CA to certifi's cacert.pem or set REQUESTS_CA_BUNDLE
       ```
     - **Node.js:**
       ```bash
       export NODE_EXTRA_CA_CERTS=/path/to/observer-ca.pem
       ```
     - **System-wide:**
       - On Ubuntu: `sudo cp observer-ca.pem /usr/local/share/ca-certificates/ && sudo update-ca-certificates`
       - On Mac: `sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain observer-ca.pem`

4. **Run your AI agent as usual.**
   - All HTTP/HTTPS traffic will be transparently captured and signaled to your backend.

---

## 🛠️ Example Docker Compose

```yaml
services:
  observer:
    build: .
    container_name: observer
    ports:
      - "8888:8888"
      - "8443:8443"
    environment:
      - BACKEND_URL=https://your-webhook-url
    volumes:
      - ./certs:/app/certs
    networks:
      - observer-net
    restart: unless-stopped

  my-ai-agent:
    image: your-ai-agent-image
    environment:
      - HTTP_PROXY=http://observer:8888
      - HTTPS_PROXY=http://observer:8443
      - NO_PROXY=localhost,127.0.0.1
    networks:
      - observer-net
    depends_on:
      - observer

networks:
  observer-net:
    driver: bridge
```

---

## 🧪 Testing

- **Test with any HTTPS endpoint:**
  ```bash
  curl -X GET https://httpbin.org/json --proxy http://localhost:8443
  ```
- **Test with your AI agent:**
  - Run your agent as usual. All outbound HTTP/HTTPS requests will be captured.
- **Check your backend/webhook:**
  - You should receive a signal for every request, with metadata and payload.

---

## 🔒 Certificate Trust (for HTTPS Interception)

- The observer uses a MITM CA certificate to decrypt HTTPS traffic.
- You must trust this CA in your AI agent's environment for seamless interception.
- See the `certs/` directory for the CA cert and instructions above.

---

## 🧩 Compatibility Table

| Agent/SDK           | Works Out of Box? | Notes                                 |
|---------------------|-------------------|---------------------------------------|
| Python requests     | ✅                | Set proxy env vars                    |
| OpenAI SDK (Python) | ✅                | Set proxy env vars                    |
| Node.js fetch/axios | ✅                | Set proxy env vars                    |
| WebSocket (wss://)  | ⚠️                | Handshake captured, messages: partial |
| gRPC                | ❌                | Needs protocol support                |

---

## 🧰 Troubleshooting

- **Port already in use?**
  - Make sure no other process is using 8888 or 8443.
  - Use `docker-compose down` before restarting.
- **SSL errors in your agent?**
  - Make sure the observer's CA is trusted by your agent.
- **No signals at your backend?**
  - Check observer logs for errors.
  - Confirm `BACKEND_URL` is set correctly.
- **WebSocket/gRPC not captured?**
  - Only HTTP/HTTPS is fully supported out of the box. Contact us for advanced protocol support.

---

## 📝 How It Works

- The observer sidecar intercepts all HTTP/HTTPS traffic from your agent.
- Every request/response is logged and a signal is sent to your backend/webhook.
- No code changes are needed in your agent—just set the proxy env vars and trust the CA.

---

## 🤝 Contributing & Support

- **Issues:** Create an issue on GitHub
- **Docs:** See the docs folder
- **Contact:** PRs and feature requests welcome!

---

**Axom Observer: Effortless, universal traffic capture for AI and beyond.**

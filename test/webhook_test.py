#!/usr/bin/env python3
"""
Test script to show exactly what's being sent to the webhook
"""

import requests
import json

def test_webhook_request():
    """Test the webhook with the exact same request the observer sends"""
    
    # Your NEW agent credentials
    agent_secret = "eMk6oyUTl8dbXLr2vyOO6sUHS3Ey7JO1"
    client_id = "7994a6a3-d70d-4976-a9fd-bf2036e7dcbd"
    webhook_url = "http://localhost:8080/api/v1/webhook/signals"
    
    # Sample signal data (what the observer would send)
    signal_data = [
        {
            "id": "signal_1735081468123456789",
            "customer_id": "t1",
            "agent_id": "1bb0f671-34c3-40b3-9045-365b471772f9",
            "timestamp": "2025-10-25T21:24:28Z",
            "protocol": "http",
            "latency_ms": 112.0,
            "metadata": {
                "provider": "OpenAI",
                "model": "gpt-4",
                "total_tokens": 25,
                "operation": "chat_completion",
                "endpoint": "/v1/chat/completions"
            },
            "source": {
                "ip": "127.0.0.1",
                "port": 0
            },
            "destination": {
                "ip": "localhost",
                "port": 443
            },
            "operation": "chat_completion",
            "status": 200
        }
    ]
    
    print("🔍 Testing Webhook Request")
    print("=" * 50)
    print(f"URL: {webhook_url}")
    print(f"Agent Secret: {agent_secret}")
    print(f"Client ID: {client_id}")
    print()
    
    # Test with Authorization Bearer header
    print("📤 Request 1: Authorization Bearer")
    print("-" * 30)
    headers = {
        "Authorization": f"Bearer {agent_secret}",
        "X-Client-ID": client_id,
        "Content-Type": "application/json"
    }
    
    print(f"Headers: {json.dumps(headers, indent=2)}")
    print(f"Body: {json.dumps(signal_data, indent=2)}")
    print()
    
    try:
        response = requests.post(webhook_url, headers=headers, json=signal_data)
        print(f"Response Status: {response.status_code}")
        print(f"Response Headers: {dict(response.headers)}")
        print(f"Response Body: {response.text}")
    except Exception as e:
        print(f"Error: {e}")
    
    print("\n" + "=" * 50)
    
    # Test with X-API-Key header
    print("📤 Request 2: X-API-Key")
    print("-" * 30)
    headers = {
        "X-API-Key": agent_secret,
        "X-Client-ID": client_id,
        "Content-Type": "application/json"
    }
    
    print(f"Headers: {json.dumps(headers, indent=2)}")
    print(f"Body: {json.dumps(signal_data, indent=2)}")
    print()
    
    try:
        response = requests.post(webhook_url, headers=headers, json=signal_data)
        print(f"Response Status: {response.status_code}")
        print(f"Response Headers: {dict(response.headers)}")
        print(f"Response Body: {response.text}")
    except Exception as e:
        print(f"Error: {e}")
    
    print("\n" + "=" * 50)
    
    # Test with X-Agent-Secret header
    print("📤 Request 3: X-Agent-Secret")
    print("-" * 30)
    headers = {
        "X-Agent-Secret": agent_secret,
        "X-Client-ID": client_id,
        "Content-Type": "application/json"
    }
    
    print(f"Headers: {json.dumps(headers, indent=2)}")
    print(f"Body: {json.dumps(signal_data, indent=2)}")
    print()
    
    try:
        response = requests.post(webhook_url, headers=headers, json=signal_data)
        print(f"Response Status: {response.status_code}")
        print(f"Response Headers: {dict(response.headers)}")
        print(f"Response Body: {response.text}")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test_webhook_request()

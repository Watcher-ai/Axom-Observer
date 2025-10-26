#!/usr/bin/env python3
"""
Mock AI Agent for Testing Axom Observer
Simulates real AI agent behavior with OpenAI, Claude, and other AI providers
"""

import requests
import json
import time
import random
from typing import Dict, Any

class MockAIAgent:
    def __init__(self, observer_url: str = "http://localhost:8888"):
        self.observer_url = observer_url
        self.session = requests.Session()
        
        # Mock API keys for different providers
        self.api_keys = {
            "openai": "sk-test-openai-key-123456789",
            "anthropic": "sk-ant-test-anthropic-key-123456789",
            "google": "AIzaSyTest-google-key-123456789",
            "cohere": "test-cohere-key-123456789"
        }
        
        print("🤖 Mock AI Agent initialized")
        print(f"📡 Observer URL: {observer_url}")
        print("🔑 Available API keys:", list(self.api_keys.keys()))

    def make_openai_request(self, endpoint: str, data: Dict[str, Any]) -> Dict[str, Any]:
        """Make OpenAI-style API request through observer"""
        url = f"{self.observer_url}{endpoint}"
        headers = {
            "Authorization": f"Bearer {self.api_keys['openai']}",
            "Content-Type": "application/json"
        }
        
        print(f"🔄 Making OpenAI request to {endpoint}")
        response = self.session.post(url, headers=headers, json=data)
        
        print(f"📊 Response: {response.status_code}")
        if response.status_code == 200:
            return response.json()
        else:
            print(f"❌ Error: {response.text}")
            return {}

    def make_anthropic_request(self, endpoint: str, data: Dict[str, Any]) -> Dict[str, Any]:
        """Make Anthropic-style API request through observer"""
        url = f"{self.observer_url}{endpoint}"
        headers = {
            "x-api-key": self.api_keys['anthropic'],
            "Content-Type": "application/json",
            "anthropic-version": "2023-06-01"
        }
        
        print(f"🔄 Making Anthropic request to {endpoint}")
        response = self.session.post(url, headers=headers, json=data)
        
        print(f"📊 Response: {response.status_code}")
        if response.status_code == 200:
            return response.json()
        else:
            print(f"❌ Error: {response.text}")
            return {}

    def test_chat_completion(self):
        """Test OpenAI chat completion"""
        print("\n🧠 Testing OpenAI Chat Completion")
        print("=" * 50)
        
        data = {
            "model": "gpt-4",
            "messages": [
                {"role": "user", "content": "Hello, can you help me write a Python function?"}
            ],
            "max_tokens": 150,
            "temperature": 0.7
        }
        
        result = self.make_openai_request("/v1/chat/completions", data)
        return result

    def test_text_completion(self):
        """Test OpenAI text completion"""
        print("\n📝 Testing OpenAI Text Completion")
        print("=" * 50)
        
        data = {
            "model": "text-davinci-003",
            "prompt": "Write a simple Python function to calculate fibonacci numbers:",
            "max_tokens": 100,
            "temperature": 0.5
        }
        
        result = self.make_openai_request("/v1/completions", data)
        return result

    def test_embeddings(self):
        """Test OpenAI embeddings"""
        print("\n🔗 Testing OpenAI Embeddings")
        print("=" * 50)
        
        data = {
            "model": "text-embedding-ada-002",
            "input": "This is a test sentence for embedding generation."
        }
        
        result = self.make_openai_request("/v1/embeddings", data)
        return result

    def test_anthropic_messages(self):
        """Test Anthropic messages API"""
        print("\n🤖 Testing Anthropic Messages")
        print("=" * 50)
        
        data = {
            "model": "claude-3-sonnet-20240229",
            "max_tokens": 100,
            "messages": [
                {"role": "user", "content": "What is the capital of France?"}
            ]
        }
        
        result = self.make_anthropic_request("/v1/messages", data)
        return result

    def test_google_ai(self):
        """Test Google AI API"""
        print("\n🔍 Testing Google AI")
        print("=" * 50)
        
        url = f"{self.observer_url}/v1beta/models/gemini-pro:generateContent"
        headers = {
            "Authorization": f"Bearer {self.api_keys['google']}",
            "Content-Type": "application/json"
        }
        
        data = {
            "contents": [{
                "parts": [{
                    "text": "Explain quantum computing in simple terms"
                }]
            }]
        }
        
        print(f"🔄 Making Google AI request")
        response = self.session.post(url, headers=headers, json=data)
        print(f"📊 Response: {response.status_code}")
        
        if response.status_code == 200:
            return response.json()
        else:
            print(f"❌ Error: {response.text}")
            return {}

    def test_cohere_generate(self):
        """Test Cohere generate API"""
        print("\n🎯 Testing Cohere Generate")
        print("=" * 50)
        
        url = f"{self.observer_url}/v1/generate"
        headers = {
            "Authorization": f"Bearer {self.api_keys['cohere']}",
            "Content-Type": "application/json"
        }
        
        data = {
            "model": "command",
            "prompt": "Write a haiku about programming:",
            "max_tokens": 50,
            "temperature": 0.7
        }
        
        print(f"🔄 Making Cohere request")
        response = self.session.post(url, headers=headers, json=data)
        print(f"📊 Response: {response.status_code}")
        
        if response.status_code == 200:
            return response.json()
        else:
            print(f"❌ Error: {response.text}")
            return {}

    def run_comprehensive_test(self):
        """Run all tests in sequence"""
        print("🚀 Starting Comprehensive AI Agent Test")
        print("=" * 60)
        print("This will test various AI providers through the observer")
        print("Signals should be sent to your local portal webhook")
        print("=" * 60)
        
        tests = [
            ("OpenAI Chat Completion", self.test_chat_completion),
            ("OpenAI Text Completion", self.test_text_completion),
            ("OpenAI Embeddings", self.test_embeddings),
            ("Anthropic Messages", self.test_anthropic_messages),
            ("Google AI", self.test_google_ai),
            ("Cohere Generate", self.test_cohere_generate),
        ]
        
        results = {}
        
        for test_name, test_func in tests:
            try:
                print(f"\n⏳ Running {test_name}...")
                result = test_func()
                results[test_name] = "✅ Success" if result else "❌ Failed"
                
                # Add delay between requests
                time.sleep(1)
                
            except Exception as e:
                print(f"❌ {test_name} failed with error: {e}")
                results[test_name] = f"❌ Error: {str(e)}"
        
        print("\n📊 Test Results Summary")
        print("=" * 30)
        for test_name, result in results.items():
            print(f"{test_name}: {result}")
        
        print("\n🔍 Check your portal webhook logs to see if signals were received!")
        return results

def main():
    """Main function to run the mock AI agent"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Mock AI Agent for testing Axom Observer")
    parser.add_argument("--observer-url", default="http://localhost:8888", 
                       help="Observer URL (default: http://localhost:8888)")
    parser.add_argument("--test", choices=["all", "openai", "anthropic", "google", "cohere"], 
                       default="all", help="Which tests to run")
    
    args = parser.parse_args()
    
    agent = MockAIAgent(args.observer_url)
    
    if args.test == "all":
        agent.run_comprehensive_test()
    elif args.test == "openai":
        agent.test_chat_completion()
        agent.test_text_completion()
        agent.test_embeddings()
    elif args.test == "anthropic":
        agent.test_anthropic_messages()
    elif args.test == "google":
        agent.test_google_ai()
    elif args.test == "cohere":
        agent.test_cohere_generate()

if __name__ == "__main__":
    main()

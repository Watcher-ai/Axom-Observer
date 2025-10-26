package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/v1/chat/completions", handleChatCompletions)
	http.HandleFunc("/v1/completions", handleCompletions)
	http.HandleFunc("/v1/embeddings", handleEmbeddings)
	http.HandleFunc("/v1/models", handleModels)
	
	log.Println("🤖 Mock AI Provider starting on port 9999")
	log.Println("📡 Available endpoints:")
	log.Println("  - POST /v1/chat/completions")
	log.Println("  - POST /v1/completions") 
	log.Println("  - POST /v1/embeddings")
	log.Println("  - GET  /v1/models")
	
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Fatal("Mock AI Provider failed:", err)
	}
}

func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	log.Printf("📨 Received chat completions request")
	
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
	
	response := map[string]interface{}{
		"id": "chatcmpl-test123",
		"object": "chat.completion",
		"created": time.Now().Unix(),
		"model": "gpt-4",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role": "assistant",
					"content": "This is a mock response from the AI provider.",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens": 10,
			"completion_tokens": 15,
			"total_tokens": 25,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCompletions(w http.ResponseWriter, r *http.Request) {
	log.Printf("📨 Received completions request")
	
	time.Sleep(50 * time.Millisecond)
	
	response := map[string]interface{}{
		"id": "cmpl-test123",
		"object": "text_completion",
		"created": time.Now().Unix(),
		"model": "gpt-3.5-turbo",
		"choices": []map[string]interface{}{
			{
				"text": "This is a mock completion response.",
				"index": 0,
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens": 5,
			"completion_tokens": 10,
			"total_tokens": 15,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	log.Printf("📨 Received embeddings request")
	
	time.Sleep(75 * time.Millisecond)
	
	response := map[string]interface{}{
		"object": "list",
		"data": []map[string]interface{}{
			{
				"object": "embedding",
				"index": 0,
				"embedding": []float64{0.1, 0.2, 0.3, 0.4, 0.5},
			},
		},
		"model": "text-embedding-ada-002",
		"usage": map[string]interface{}{
			"prompt_tokens": 5,
			"total_tokens": 5,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleModels(w http.ResponseWriter, r *http.Request) {
	log.Printf("📨 Received models request")
	
	response := map[string]interface{}{
		"object": "list",
		"data": []map[string]interface{}{
			{
				"id": "gpt-4",
				"object": "model",
				"created": 1677610602,
				"owned_by": "openai",
			},
			{
				"id": "gpt-3.5-turbo",
				"object": "model", 
				"created": 1677610602,
				"owned_by": "openai",
			},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

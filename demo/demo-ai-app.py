from flask import Flask, request, jsonify
import time
import uuid
from datetime import datetime

app = Flask(__name__)

@app.route('/v1/chat/completions', methods=['POST'])
def chat_completions():
    data = request.get_json(silent=True) or {}
    model = data.get('model', 'gpt-4')
    messages = data.get('messages', [])
    max_tokens = data.get('max_tokens', 100)
    temperature = data.get('temperature', 0.7)
    
    # Extract user message
    user_message = ""
    for msg in messages:
        if msg.get('role') == 'user':
            user_message = msg.get('content', '')
            break
    
    # Simulate processing time
    time.sleep(0.1)
    
    # Simulate token usage
    input_tokens = len(user_message.split()) * 1.3
    output_tokens = min(max_tokens, len(user_message.split()) * 2)
    
    return jsonify({
        'id': f"chatcmpl-{uuid.uuid4().hex[:8]}",
        'object': 'chat.completion',
        'created': int(datetime.utcnow().timestamp()),
        'model': model,
        'choices': [{
            'index': 0,
            'message': {
                'role': 'assistant',
                'content': f"Simulated response to: {user_message}"
            },
            'finish_reason': 'stop'
        }],
        'usage': {
            'prompt_tokens': int(input_tokens),
            'completion_tokens': int(output_tokens),
            'total_tokens': int(input_tokens + output_tokens)
        }
    })

@app.route('/v1/embeddings', methods=['POST'])
def embeddings():
    data = request.get_json(silent=True) or {}
    model = data.get('model', 'text-embedding-ada-002')
    input_text = data.get('input', '')
    
    # Handle both string and list inputs
    if isinstance(input_text, str):
        texts = [input_text]
    else:
        texts = input_text
    
    # Simulate processing time
    time.sleep(0.05)
    
    embeddings = []
    for text in texts:
        tokens = len(text.split()) * 1.3
        embeddings.append({
            'object': 'embedding',
            'embedding': [0.1] * 1536,  # Simulated embedding
            'index': len(embeddings)
        })
    
    return jsonify({
        'object': 'list',
        'data': embeddings,
        'model': model,
        'usage': {
            'prompt_tokens': int(sum(len(text.split()) * 1.3 for text in texts)),
            'total_tokens': int(sum(len(text.split()) * 1.3 for text in texts))
        }
    })

@app.route('/v1/models', methods=['GET'])
def models():
    return jsonify({
        'object': 'list',
        'data': [
            {
                'id': 'gpt-3.5-turbo',
                'object': 'model',
                'created': 1677610602,
                'owned_by': 'openai'
            },
            {
                'id': 'gpt-4',
                'object': 'model',
                'created': 1677610602,
                'owned_by': 'openai'
            },
            {
                'id': 'text-embedding-ada-002',
                'object': 'model',
                'created': 1677610602,
                'owned_by': 'openai'
            }
        ]
    })

@app.route('/error', methods=['GET'])
def error():
    return jsonify({'error': 'something went wrong'}), 500

@app.route('/slow', methods=['GET'])
def slow():
    time.sleep(2)
    return jsonify({'result': 'slow response'})

@app.route('/status/<int:code>', methods=['GET'])
def status(code):
    return jsonify({'status': code}), code

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5002)

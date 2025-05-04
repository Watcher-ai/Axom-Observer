from flask import Flask, request, jsonify
import time

app = Flask(__name__)

@app.route('/predict', methods=['GET', 'POST'])
def predict():
    # Simulate latency
    time.sleep(0.1)
    data = request.get_json(silent=True) or {}
    return jsonify({'result': 'success', 'input': data})

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
    app.run(host='0.0.0.0', port=5000)

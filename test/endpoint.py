import os
from flask import Flask, jsonify, request
from functools import wraps

app = Flask(__name__)

@app.route('/', methods=['GET'])
def list_commands():
    return jsonify([
        {"id": "command1", "command": "ls -la"},
        {"id": "command2", "command": "df -h"}
    ])

@app.route('/commands/<command_id>', methods=['GET'])
def get_command(command_id):
    commands = {
        "command1": {"id": "command1", "command": "ls -la"},
        "command2": {"id": "command2", "command": "df -h"}
    }
    return jsonify(commands.get(command_id, {}))

@app.route('/results', methods=['POST'])
def receive_command_result():
    data = request.json
    command_id = data.get('id')
    result = data.get('result')

    print(f"Received result for command {command_id}")
    print(f"Result: {result}")

    return jsonify({"status": "ok"}), 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)

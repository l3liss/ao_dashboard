import json

def load_state(file_path):
    try:
        with open(file_path, 'r') as f:
            return json.load(f)
    except Exception as e:
        print(f"Failed to load state.json: {e}")
        return {}

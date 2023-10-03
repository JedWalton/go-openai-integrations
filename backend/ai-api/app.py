from flask import Flask, request, jsonify, abort
import os
from dotenv import load_dotenv

# NLTK
import nltk
nltk.download('stopwords')
from nltk.tokenize.texttiling import TextTilingTokenizer

import re

load_dotenv()

SECRET_KEY = os.getenv("X_AI_API_KEY")
print(f"Loaded SECRET_KEY from .env: {SECRET_KEY}")  # Diagnostic print


app = Flask(__name__)


@app.route('/split_sentences', methods=['POST'])
def split_sentences():
    try:
        secret_key = request.headers.get('X-AI-API-KEY')
        print(f"Received X_AI_API_KEY header: {secret_key}")  # Diagnostic print

        if not secret_key or secret_key != SECRET_KEY:
            print("Unauthorized due to mismatched or missing secret key.")  # Diagnostic print
            abort(401, description="Unauthorized")

        text = request.json.get('text', "")
        text = sanitize_input(text)
        if not text:
            return jsonify({"error": "No text provided"}), 400

        tt = TextTilingTokenizer()
        try:
            segments = tt.tokenize(text)
        except ValueError as e:
            if str(e) == "No paragraph breaks were found(text too short perhaps?)":
                segments = [text]  # Return the same text as a single segment
            else:
                raise e  # If it's a different error, raise it to be caught by the outer exception handler

        return jsonify(segments)
    except Exception as e:
        print(f"Exception occurred: {str(e)}")  # Diagnostic print
        return jsonify({"error": str(e)}), 500

def sanitize_input(text):
    """Sanitize the input text by removing control characters and ensuring it's a string."""
    if not isinstance(text, str):
        return "Invalid input. Expected a string."
    
    # Remove control characters
    sanitized_text = re.sub(r'[\x00-\x1f\x7f-\x9f]', ' ', text)
    
    # Replace multiple spaces with a single space
    sanitized_text = re.sub(r'\s+', ' ', sanitized_text).strip()

    return sanitized_text

@app.errorhandler(500)
def internal_error(error):
    app.logger.error('Server Error: %s', (error))
    return jsonify({"error": "Internal server error", "details": str(error)}), 500

@app.errorhandler(404)
def not_found(error):
    app.logger.error('Not Found: %s', (error))
    return "404 error"

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)


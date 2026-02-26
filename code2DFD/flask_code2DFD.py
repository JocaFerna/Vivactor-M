from flask import Flask, jsonify, request

import code2DFD

import os
from dotenv import load_dotenv
from flask_cors import CORS

load_dotenv()

# Initialize ONLY ONCE
app = Flask(__name__, instance_relative_config=True)

# Apply CORS to the instance we are actually using
# Adding methods ensures OPTIONS preflight requests work correctly
CORS(app, resources={r"/*": {"origins": "http://localhost:5173"}}, 
     methods=["GET", "POST", "OPTIONS"])


@app.get('/')
def index():
    index_message = ("API for DFD extraction. \
    Provide a GitHub URL to endpoint /dfd as parameter \"url\" to receive the extracted DFD: \
    /dfd?url=https://github.com/georgwittberger/apache-spring-boot \
           -microservice-example; \
                     Optionally provide a commit hash as \"commit\" parameter")

    return index_message


@app.get('/dfd_local')
def dfd_local():
    
    url = request.args.get("url")

    if not url:
        return "Please specify a local URL, e.g. /dfd_local?url=repository_to_analyse/piggymetrics "

    # Call Code2DFD
    results = code2DFD.api_invocation(url, None)

    # Create response JSON object and return it
    response = jsonify(**results)

    return response


@app.get('/dfd')
def dfd():

    url = request.args.get("url")
    commit = request.args.get("commit", None)

    if not url:
        return "Please specify a URL, e.g. /dfd?url=https://github.com/georgwittberger/apache-spring-boot" \
               "-microservice-example "

    # Call Code2DFD
    results = code2DFD.api_invocation(url, commit)

    # Create response JSON object and return it
    response = jsonify(**results)

    return response


# starts local server
if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=int(os.getenv("PORT", 5000)))

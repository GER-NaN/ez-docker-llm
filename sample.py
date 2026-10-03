"""Three example calls. Run with: python sample.py"""
import json
import urllib.error
import urllib.request

URL = "http://127.0.0.1:8085/prompt"
MODEL = "qwen-9b"

# Example context provided to the model
TEXT = "Mayor Bob Barker discussed the financial situation with treasurer Jane Janet"

# A JSON schema: a list of objects, each with a name and a role.
# In this sample script, we ask our model to return answers in this format.
FORMAT = {
    "type": "array",
    "items": {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "role": {"type": "string"},
        },
        "required": ["name", "role"],
    },
}


def ask(body):
    request = urllib.request.Request(
        URL,
        data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json"},
    )
    try:
        response = urllib.request.urlopen(request, timeout=900)
        return response.read().decode()
    except urllib.error.HTTPError as error:
        return error.read().decode()

# A simple question with no context
print("1. prompt")
print(ask({
    "model": MODEL,
    "prompt": "Who was the president in 2006?",
}))

# A simple question with a text for it to use
print("\n2. prompt and text")
print(ask({
    "model": MODEL,
    "prompt": "Read the news article attached and answer the question, who works at city hall",
    "text": TEXT,
}))

# A simple question with a text and a format for the answer
print("\n3. prompt, text and format")
print(ask({
    "model": MODEL,
    "prompt": "Read the news article attached and answer the question along with their name and role, Who works at city hall",
    "text": TEXT,
    "format": FORMAT,
}))

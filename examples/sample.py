"""Three example calls. Run with: python examples/sample.py"""
import json
import urllib.error
import urllib.request

URL = "http://127.0.0.1:8085/prompt"
MODEL = "qwen-9b"

# The text we give the model to read.
TEXT = "Mayor Bob Barker discussed the financial situation with treasurer Jane Janet"

# The format we want the answer in, as a JSON schema:
# a list of objects, each with a name and a role.
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


# A question on its own.
print("1. prompt")
print(ask({
    "model": MODEL,
    "prompt": "Who was the president in 2006?",
}))

# A question about a text.
print("\n2. prompt and text")
print(ask({
    "model": MODEL,
    "prompt": "Read the news article attached and answer the question, who works at city hall",
    "text": TEXT,
}))

# The same question, with the answer in our format.
print("\n3. prompt, text and format")
print(ask({
    "model": MODEL,
    "prompt": "Read the news article attached and answer the question along with their name and role, Who works at city hall",
    "text": TEXT,
    "format": FORMAT,
}))

# ez-docker-llm

Local LLMs in Docker with a small HTTP API.

This is more than the bare minimum to get a model running. It is the bare minimum that makes a model easy to call from a script. See `examples/sample.py`.

You need Docker with NVIDIA GPU support, and Python 3 to run the sample.

## Quick start

Download a model (5.7 GB):

```
cd models
curl -L -O https://huggingface.co/unsloth/Qwen3.5-9B-GGUF/resolve/main/Qwen3.5-9B-Q4_K_M.gguf
cd ..
```

Start it:

```
docker compose up -d --build
```

Try it:

```
python examples/sample.py
```

Stop it:

```
docker compose down
```

## API

The API is at `http://127.0.0.1:8085`.

**GET /health**

```json
{"status": "ok", "llama": "ok", "models": 4, "loaded": ["qwen-9b"]}
```

**GET /models**

```json
[{"name": "qwen-9b", "loaded": true}, {"name": "qwen-4b", "loaded": false}]
```

**POST /prompt**

| Field | |
|---|---|
| `model` | Model name |
| `prompt` | Your question or instruction |
| `text` | Optional. Something for the model to read |
| `format` | Optional. A JSON schema for the answer |

A question:

```json
{"model": "qwen-9b", "prompt": "Who was the president in 2006?"}
```
```json
{"model": "qwen-9b", "answer": "In 2006, George W. Bush was the President of the United States.", "seconds": 1.2}
```

A question about a text:

```json
{
  "model": "qwen-9b",
  "prompt": "Read the news article attached and answer the question, who works at city hall",
  "text": "Mayor Bob Barker discussed the financial situation with treasurer Jane Janet"
}
```
```json
{"model": "qwen-9b", "answer": "Mayor Bob Barker and Treasurer Jane Janet work at City Hall.", "seconds": 0.41}
```

The same, with the answer as JSON:

```json
{
  "model": "qwen-9b",
  "prompt": "Read the news article attached and answer the question along with their name and role, Who works at city hall",
  "text": "Mayor Bob Barker discussed the financial situation with treasurer Jane Janet",
  "format": {
    "type": "array",
    "items": {
      "type": "object",
      "properties": {"name": {"type": "string"}, "role": {"type": "string"}},
      "required": ["name", "role"]
    }
  }
}
```
```json
{"model": "qwen-9b", "answer": [{"name": "Bob Barker", "role": "Mayor"}, {"name": "Jane Janet", "role": "Treasurer"}], "seconds": 0.74}
```

The first prompt to a model is slow while it loads. One model is loaded at a time. Asking for a different one swaps it in.

## Add your own model

1. Put the `.gguf` file in `models/`.
2. Add an entry to `models.ini`. The name in brackets is the name you use in the API.

```ini
[my-model]
model = /models/my-model.gguf
```

3. Restart:

```
docker compose up -d --force-recreate
```

If a prompt returns `failed to load`, that model's file is not in `models/`.

### Customize a model

Add settings under its entry. For example, `temp = 0` makes the model give the same answer every time:

```ini
[my-model]
model = /models/my-model.gguf
temp = 0
```

Common settings:

| Setting | Default here | |
|---|---|---|
| `temp` | `0.8` | Randomness. `0` gives the same answer every time |
| `ctx-size` | `16384` | How much text a prompt can hold, in tokens |
| `n-predict` | no limit | Longest answer, in tokens |
| `reasoning-budget` | `0` | `0` turns thinking off, `-1` lets the model think |
| `n-gpu-layers` | `99` | How much of the model runs on the GPU. `99` is all of it |

Settings under `[*]` apply to every model. Any [llama-server option](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md) works, written without the leading dashes.

### Included models

`models.ini` comes with four entries. Download the ones you want:

| Name | Size | Download |
|---|---|---|
| `qwen-4b` | 2.7 GB | https://huggingface.co/unsloth/Qwen3.5-4B-GGUF/resolve/main/Qwen3.5-4B-Q4_K_M.gguf |
| `qwen-9b` | 5.7 GB | https://huggingface.co/unsloth/Qwen3.5-9B-GGUF/resolve/main/Qwen3.5-9B-Q4_K_M.gguf |
| `gemma-e4b` | 5.0 GB | https://huggingface.co/unsloth/gemma-4-E4B-it-GGUF/resolve/main/gemma-4-E4B-it-Q4_K_M.gguf |
| `gemma-26b` | 13.6 GB | https://huggingface.co/unsloth/gemma-4-26B-A4B-it-GGUF/resolve/main/gemma-4-26B-A4B-it-UD-IQ4_XS.gguf |

## Settings

Optional. To change one, create a `.env` file next to `docker-compose.yml`, for example `HOST_PORT=9000`.

| Setting | Default | |
|---|---|---|
| `MODELS_DIR` | `./models` | Where the model files are |
| `HOST_PORT` | `8085` | Port on your machine |
| `MODELS_MAX` | `1` | Models loaded at once |

## License

Public domain. See `LICENSE`.

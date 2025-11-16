<pre>
██████╗ ███████╗██╗     ██╗ ██████╗
██╔══██╗██╔════╝██║     ██║██╔════╝
██████╔╝█████╗  ██║     ██║██║
██╔══██╗██╔══╝  ██║     ██║██║
██║  ██║███████╗███████╗██║╚██████╗
╚═╝  ╚═╝╚══════╝╚══════╝╚═╝ ╚═════╝
</pre>

[![CI](https://img.shields.io/github/actions/workflow/status/ju4n97/relic/ci.yaml?branch=main&style=flat-square)](https://github.com/ju4n97/relic/actions/workflows/ci.yaml)
![GitHub Release](https://img.shields.io/github/v/release/ju4n97/relic?style=flat-square&include_prereleases)
[![Go Report Card](https://goreportcard.com/badge/github.com/ju4n97/relic?style=flat-square)](https://goreportcard.com/report/github.com/ju4n97/relic)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?style=flat-square)](https://pkg.go.dev/github.com/ju4n97/relic/sdk-go)

> [!IMPORTANT]
> In active development. API design driven by production voice assistant use cases I currently have, but built for general-purpose AI applications.

RELIC is a local AI runtime that lets you run LLMs, speech-to-text, text-to-speech, vision models, and embeddings through a single HTTP/gRPC API. Useful for building voice assistants, chatbots, or any AI-powered application without cloud dependencies.

## Quick start

### CPU

```bash
docker run -p 8080:8080 -p 50051:50051 ghcr.io/ju4n97/relic:latest
```

### NVIDIA GPU

Requires [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html).

```bash
docker run -p 8080:8080 -p 50051:50051 --gpus all ghcr.io/ju4n97/relic:cuda
```

## Configuration

RELIC uses a `relic.yaml` file to define which models to download and which services to expose. Models are downloaded automatically from the specified source on the first run.

```yaml
# relic.yaml
version: "1"

models:
    llama-cpp-qwen2.5-1.5b-instruct:
        type: llm
        backend: llama.cpp
        source:
            huggingface:
                repo: Qwen/Qwen2.5-1.5B-Instruct-GGUF
                include: ["qwen2.5-1.5b-instruct-q4_k_m.gguf"]

    whisper-cpp-small:
        type: stt
        backend: whisper.cpp
        source:
            huggingface:
                repo: ggerganov/whisper.cpp
                include: ["ggml-small.bin"]
        tags: [multilingual, streaming]

    piper-es-ar-daniela:
        type: tts
        backend: piper
        source:
            huggingface:
                repo: rhasspy/piper-voices
                include: ["es/es_AR/daniela/high/*"]
        tags: [spanish, argentina, high-quality]

services:
    llm:
        models: [llama-cpp-qwen2.5-1.5b-instruct]
    stt:
        models: [whisper-cpp-small]
    tts:
        models: [piper-es-ar-daniela]
```

### Environment variables

| Variable                   | Description                               |
| -------------------------- | ----------------------------------------- |
| `RELIC_ENV`              | Runtime environment (`dev`, `prod`, etc.) |
| `RELIC_SERVER_HTTP_PORT` | HTTP server port                          |
| `RELIC_SERVER_GRPC_PORT` | gRPC server port                          |
| `RELIC_MODELS_PATH`      | Path to models directory                  |
| `RELIC_CONFIG_PATH`      | Path to config file (`relic.yaml`)      |

## Examples

Working demos can be found in the [examples](examples) directory.

### Go SDK examples

- [Basic completion](examples/go-basic-completion): Simple LLM inference
- [Streaming response](examples/go-streaming-response): Real-time token streaming
- [Multi-turn conversation](examples/go-multi-turn-conversation) - Stateful chat
- [Speech-to-Text](examples/go-speech-to-text): Audio transcription
- [Text-to-Speech](examples/go-text-to-speech): Audio synthesis
- [Voice assistant pipeline](examples/go-voice-assistant-pipeline): Complete STT -> LLM -> TTS flow

## Supported backends

### 🟡 Experimental

| Type    | Backend                                                 | Source                               | Acceleration    | License | Notes                                     |
| ------- | ------------------------------------------------------- | ------------------------------------ | --------------- | ------- | ----------------------------------------- |
| **LLM** | [llama.cpp](https://github.com/ggml-org/llama.cpp)      | [`backend/llama`](backend/llama)     | CPU, CUDA 11/12 | MIT     | Qwen, Mistral, Llama, Phi, DeepSeek, etc. |
| **STT** | [whisper.cpp](https://github.com/ggerganov/whisper.cpp) | [`backend/whisper`](backend/whisper) | CPU, CUDA 12    | MIT     | All Whisper variants (tiny to large-v3)   |
| **TTS** | [Piper](https://github.com/rhasspy/piper)               | [`backend/piper`](backend/piper)     | CPU             | MIT     | 200+ voices across 50+ languages          |

## Roadmap

Planned backends for future releases:

| Type           | Backend                                                                  | License    | Description                                 | Status     |
| -------------- | ------------------------------------------------------------------------ | ---------- | ------------------------------------------- | ---------- |
| **STT**        | [Vosk](https://github.com/alphacep/vosk-api)                             | Apache 2.0 | Offline speech recognition                  | 🔴 Planned |
| **NLU**        | [Rasa](https://github.com/RasaHQ/rasa)                                   | Apache 2.0 | Intent classification and entity extraction | 🔴 Planned |
| **VAD**        | [Silero VAD](https://github.com/snakers4/silero-vad)                     | MIT        | Voice activity detection                    | 🔴 Planned |
| **TTS**        | [Coqui TTS](https://github.com/coqui-ai/TTS)                             | MPL 2.0    | Neural text-to-speech                       | 🔴 Planned |
| **Vision**     | [ONNX Runtime + OpenCV](https://github.com/microsoft/onnxruntime)        | MIT        | Image processing                            | 🔴 Planned |
| **Vision**     | [Ultralytics YOLO](https://github.com/ultralytics/ultralytics)           | AGPL-3.0   | Object detection                            | 🔴 Planned |
| **Embeddings** | [sentence-transformers](https://github.com/UKPLab/sentence-transformers) | Apache 2.0 | Text embeddings                             | 🔴 Planned |
| **Embeddings** | [nomic-embed-text](https://github.com/nomic-ai/nomic)                    | Apache 2.0 | Dense embeddings                            | 🔴 Planned |

**Status legend:**

- 🟢 Supported: tested, stable, and recommended for production.
- 🟡 Experimental: functional but subject to changes, bugs, or limitations.
- 🟠 Development: active integration with features still under construction.
- 🔴 Planned: intended for future implementation (PRs welcome).

## Architecture (as of October 2025)

```mermaid
flowchart TD
    subgraph CLIENTS[External interfaces / Clients]
        A1[CLI / SDK / API]
        A2[External applications]
        A3[Third-Party agents or services]
    end

    subgraph RELIC[RELIC]
        direction TB

        subgraph CONTROL[Control layer]
            B1[Model registry / State]
            B2[gRPC and HTTP Server]
        end

        subgraph BACKENDS[Inference backends]
            C1[LLM: Qwen, Mistral, etc.]
            C2[NLU: Rasa, spaCy etc.]
            C3[STT: Whisper, Vosk, etc.]
            C4[TTS: Kokoro, Piper, etc.]
            C5[Embeddings]
            C6[Vision]
        end

        subgraph STORAGE[Storage and configuration]
            E1[Model cache]
            E2[Metadata / Config]
        end
    end

    A1 -->|Inference / Management| B2
    A2 -->|Streaming / Batch| B2
    A3 -->|Local control| B2
    B2 --> B1
    B2 --> BACKENDS
    BACKENDS --> B2
    B1 --> STORAGE
```

## Development

### Requirements

- [Go v1.25+](https://go.dev)
- [CMake v3.22+](https://cmake.org)
- [Docker](https://www.docker.com)
- [Task](https://taskfile.dev)
- [protoc](https://github.com/protocolbuffers/protobuf)

```bash
git clone --recursive https://github.com/ju4n97/relic.git
cd relic

task install
# Build backends (this may take several minutes the first time)
task build-third-party          # CPU
# task build-third-party-cuda   # CUDA
task help
```

[Taskfile.yaml](./Taskfile.yaml) is your guide.

## License

[MIT](LICENSE)

# Rag Server

This is a repo for me to learn about LLM, Embedding Model, LangChain, Golang, and Milvus (Vector Database).

## Tech stack used

- Go (Programming Language)
- OpenAI / Ollama (Large Language Model)
- Embedding Model
- LangChainGo (Framework)
- Milvus (Vector database)
- Viper (Go configuration)
- Docker (Container)
- Kubernetes (Container orchestation)
- Helm (Kubernetes package manager)
- Terraform (Infrastructure as Code)

## Prerequisites

### Install milvus

I recommend you to install either with Docker or Helm

#### With Docker

#### With Helm Chart (Kubernetes)

### Setup OpenAI API 

#### With Ollama

you could use Ollama as a free and offline alternative for OpenAI, as I did

```bash
sudo systemctl start ollama
```

or 

```bash
ollama serve
```

### Run the Server

#### With Golang

#### With Docker

#### With Kubernetes Cluster

> Make sure you have a kubernetes cluster running.
> This is an example for local cluster setup with Kubernetes in Docker (KinD)
> ```bash
> kind create cluster
> ```

**Install Milvus**

> You could use Terraform to skip this step
> ```bash
> terraform init
> terraform apply
> ```

Add Milvus Helm Chart if you haven't

```bash
helm repo add zilliztech https://zilliztech.github.io/milvus-helm/
```

Install Milvus to Kubernetes Cluster

> This might take a long time.

```bash
helm install milvus-release zilliztech/milvus \
  --create-namespace \
  --namespace milvus-namespace \
  --set image.all.tag=v2.6.6 \
  --set pulsarv3.enabled=false \
  --set woodpecker.enabled=true \
  --set streaming.enabled=true \
  --set indexNode.enabled=false
```

Activate Ollama in your host (if you use it)

```bash
sudo systemctl start ollama
```

Install model

```bash
ollama pull gemma3:1b-it-qat
ollama pull embeddinggemma:latest
```

Forward Port

```bash
kubectl port-forward -n milvus-namespace svc/milvus-release 19530:19530
```

Run the app from you host

```bash
go run ./cmd/rest
```

Output:

```txt
2025/12/04 20:43:18 Connecting to milvus database...
2025/12/04 20:43:18 Connected to milvus database
2025/12/04 20:43:18 Running database migrations...
2025/12/04 20:43:18 Database migrations completed
2025/12/04 20:43:18 Server listening on  :8080
```

## API

### Add Document

Endpoint: POST /add

```json
{
  "document": [
    "My name is yasa"
  ]
}
```

Response:

```json
{
    "status": "success"
}
```

### Retrieve documents

Endpoint: POST /query

```json
{
  "query": "what is my name?"
}
```

Response (Failed):

```json
{
    "result": [
        "I don't know.\n"
    ]
}
```

Response (Success):

```json
{
    "result": [
        "You are yasa.\n"
    ]
}
```

## TODO

- change the prompt to response in json format instead

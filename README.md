# Rag Server Demo

This is a repo for me to learn about LLM, Embedding Model, LangChain, Golang, and Milvus (Vector Database).

## Prerequisites

### Install milvus

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

## API

### Add Document

Endpoint: POST /add

```json
{
  "document": [
    "my name is yasa"
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

Response:

```txt
Your name is yasa.
```





## TODO

- change the prompt to response in json format instead

terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "3.1.1"
    }
  }
}

provider "helm" {
  kubernetes = {
    config_path = "~/.kube/config"
  }
  # Configuration options
}

resource "helm_release" "milvus" {
  name             = "milvus-release"
  repository       = "https://zilliztech.github.io/milvus-helm"
  chart            = "milvus"
  namespace        = "milvus-namespace"
  create_namespace = true

  set = [
    {
      name  = "image.all.tag"
      value = "v2.6.6"
    },
    {
      name = "cluster.enabled"
      value = false
    },
    {
      name  = "pulsarv3.enabled"
      value = false
    },
    {
      name = "standalone.messageQueue"
      value = "woodpecker"
    },
    {
      name  = "woodpecker.enabled"
      value = true
    },
    {
      name  = "streaming.enabled"
      value = true
    },
  ]
}

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.41.0"

    }
  }
  backend "gcs" {
    bucket = "com-dkisler-sys-go-mod-analysis"
    prefix = "terraform"
  }
}

locals {
  region = "us-central1"
}

provider "google" {
  project = "go-mod-analysis"
  region  = local.region
  zone    = "us-central1-c"
}

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.41.0"

    }
  }
  backend "gcs" {
    credentials = "key.json"
    bucket      = "com-dkisler-terraform"
    prefix      = "go-mod-analysis"
  }
}

locals {
  region   = "us-central1"
  projects = toset(["go-mod-analysis"])
}

provider "google" {
  project = "dkisler-root"
  region  = local.region
  zone    = "us-central1-c"
}

resource "google_project" "this" {
  for_each        = local.projects
  name            = each.key
  project_id      = each.key
  org_id          = var.org_id
  billing_account = var.billing_account
}

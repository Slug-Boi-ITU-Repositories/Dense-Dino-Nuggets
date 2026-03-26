terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
  backend "s3" {
    endpoint = "https://fra1.digitaloceanspaces.com"
    region   = "eu-fra1"          # required field, value doesn't matter for DO
    bucket   = "minitwit-bucket"
    key      = "minitwit/terraform.tfstate"

    skip_credentials_validation = true
    skip_metadata_api_check     = true
  }
}

provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_volume" "pgdata" {
  region      = "fra1"
  name        = "volume-fra1-01"
  size        = 50
  description = "Postgres data volume"
}

resource "digitalocean_droplet" "minitwit" {
  name   = "minitwit-prod"
  region = "fra1"
  size   = "s-1vcpu-1gb"
  image  = "ubuntu-22-04-x64"

  user_data = templatefile("${path.module}/provision.sh.tpl", {
    postgres_user     = var.postgres_user
    postgres_password = var.postgres_password
    postgres_db       = var.postgres_db
    volume_mount      = var.volume_mount
  })
}

resource "digitalocean_volume_attachment" "pgdata" {
  droplet_id = digitalocean_droplet.minitwit.id
  volume_id  = digitalocean_volume.pgdata.id
}
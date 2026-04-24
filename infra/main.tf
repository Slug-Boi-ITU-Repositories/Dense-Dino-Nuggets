terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
  backend "s3" {
    endpoint = "https://fra1.digitaloceanspaces.com"
    region   = "us-east-1"          # required field, value doesn't matter for DO
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
}

resource "digitalocean_droplet" "minitwit" {
  name   = "minitwit"
  region = "fra1"
  size   = "s-1vcpu-1gb"
  image  = "ubuntu-22-04-x64"

  lifecycle {
    ignore_changes = [
      image,
      user_data,
      public_networking,
    ]
  }

  user_data = templatefile("${path.module}/minitwit_provision.sh.tpl", {
    postgres_user         = var.postgres_user
    postgres_password     = var.postgres_password
    postgres_db           = var.postgres_db
    volume_mount          = var.volume_mount
    postgres_host         = var.postgres_host
    monitor_ip            = var.monitor_ip
    docker_username       = var.docker_username
    container_name_prefix = var.container_name_prefix
    postgres_port         = var.postgres_port
    db_ssl_mode           = var.db_ssl_mode
    postgres_cpu_limit    = var.postgres_cpu_limit
    postgres_mem_limit    = var.postgres_mem_limit
    app_replicas          = var.app_replicas
    app_cpu_limit         = var.app_cpu_limit
    app_mem_limit         = var.app_mem_limit
  })

  connection {
    type        = "ssh"
    user        = "root"
    private_key = file("~/.ssh/devops_rsa")
    host        = self.ipv4_address
  }

  provisioner "file" {
    source      = "./docker-compose.yml"
    destination = "/vagrant/docker-compose.yml"
  }

  provisioner "file" {
    source      = "./scripts/monitor_setup.sh"
    destination = "/vagrant/monitor_setup.sh"
  }

  provisioner "file" {
    source      = ".ssh/id_monitor"
    destination = "/tmp/id_monitor"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /vagrant/monitor_setup.sh",
    ]
  }
}

resource "digitalocean_droplet" "monitoring" {
  name   = "monitoring"
  region = "fra1"
  size   = "s-1vcpu-2gb"
  image  = "ubuntu-22-04-x64"

  lifecycle {
    ignore_changes = [
      public_networking,
      user_data
    ]
  }

  user_data = templatefile("${path.module}/monitor_provision.sh.tpl", {
      monitor_pub_key       = var.monitor_pub_key
  })

  connection {
    type        = "ssh"
    user        = "root"
    private_key = file("~/.ssh/monitor_id")
    host        = self.ipv4_address
  }

  provisioner "file" {
    source      = "./prometheus/prometheus_prod.yml"
    destination = "/prometheus/prometheus_prod.yml"
  }

  provisioner "file" {
    source      = "./loki/loki-config.yml"
    destination = "/loki/loki-config.yml"
  }

  provisioner "file" {
    source      = "./grafana/"
    destination = "/grafana/"
  }
}

resource "digitalocean_volume_attachment" "pgdata" {
  droplet_id = digitalocean_droplet.minitwit.id
  volume_id  = digitalocean_volume.pgdata.id
}
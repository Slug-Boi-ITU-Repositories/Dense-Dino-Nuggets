# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|

  config.vm.define "minitwit" do |server|
    server.vm.hostname = "minitwit"

    server.vm.provider :utm do |u, override|
      config.vm.synced_folder "./db", "/db" , owner: "root", group: "root"
      override.vm.box = "utm/bookworm"
      u.memory = 2048
      u.cpus = 2
      override.vm.provision "file", source: "./docker-compose.yml", destination: "/tmp/docker-compose.yml"
    end

    server.vm.provider :virtualbox do |vb, override|
      override.vm.box = "ubuntu/jammy64"
      vb.memory = 2048
      vb.cpus = 2
    end

    server.vm.provider :libvirt do |lv, override|
      override.vm.box = "generic/ubuntu2204"
      lv.memory = 2048
      lv.cpus = 2
      lv.driver = "kvm"
      lv.default_prefix = ""
    end

    # DigitalOcean (Cloud)
    server.vm.provider :digital_ocean do |provider, override|
      override.vm.box = "digital_ocean"
      override.vm.box_url = "https://github.com/devopsgroup-io/vagrant-digitalocean/raw/master/box/digital_ocean.box"
      provider.token = ENV["DIGITAL_OCEAN_TOKEN"]
      provider.ssh_key_name = ENV["SSH_KEY_NAME"]
      override.nfs.functional = false
      override.vm.allowed_synced_folder_types = :rsync
      override.ssh.private_key_path = '~/.ssh/new_testing'
      provider.image = "ubuntu-22-04-x64"
      provider.region = "fra1"
      provider.size = "s-1vcpu-1gb"
      provider.vm.provision "file", source: "./docker-compose.yml", destination: "/vagrant/docker-compose.yml"
    end

    server.vm.network "forwarded_port", guest: 8080, host: 8080

    server.vm.provision "shell", env: {
      "DOCKER_USERNAME" => ENV['DOCKER_USERNAME'] || "flakiator",
      "CONTAINER_NAME_PREFIX" => ENV['CONTAINER_NAME_PREFIX'],
      "POSTGRES_USER" => ENV['POSTGRES_USER'],
      "POSTGRES_PASSWORD" => ENV['POSTGRES_PASSWORD'],
      "POSTGRES_DB" => ENV['POSTGRES_DB'],
      "POSTGRES_HOST" => ENV['POSTGRES_HOST'],
      "POSTGRES_PORT" => ENV['POSTGRES_PORT'],
      "DB_SSL_MODE" => ENV['DB_SSL_MODE'],
      "VOLUME_MOUNT" => ENV['VOLUME_MOUNT'] || "/mnt/pgdata",
      "MONITOR_IP" => ENV['MONITOR_IP'],
      "MONITOR_PUB_KEY" => ENV['MONITOR_PUB_KEY'],
      "POSTGRES_CPU_LIMIT" => ENV['POSTGRES_CPU_LIMIT'] || "0.2",
      "POSTGRES_MEM_LIMIT" => ENV['POSTGRES_MEM_LIMIT'] || "300M",
      "APP_REPLICAS" => ENV['APP_REPLICAS'] || "2",
      "APP_CPU_LIMIT" => ENV['APP_CPU_LIMIT'] || "0.2",
      "APP_MEM_LIMIT" => ENV['APP_MEM_LIMIT'] || "256M"
      }, inline: <<-SHELL
      set -euo pipefail

      VOLUME_DEVICE="/dev/sda"
      IMAGE_NAME="minitwitimage"
      DOCKER_IMAGE="$DOCKER_USERNAME/$IMAGE_NAME"
      COMPOSE_FILE="docker-compose.yml"
      STACK_NAME="minitwit"

      REQUIRED_VARS="POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_HOST MONITOR_IP MONITOR_PUB_KEY"
      for var in $REQUIRED_VARS; do
        if [ -z "${!var}" ]; then
          echo "ERROR: $var is not set. Please set it in your host environment."
          exit 1
        fi
      done

      sudo apt-get update -y
      sudo apt-get install -y ca-certificates curl gnupg lsb-release

      if [ ! -f /etc/apt/keyrings/docker.gpg ]; then
        echo "Setting up Docker's GPG key and repository (first time only)..."
        sudo install -m 0755 -d /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg | sudo gpg --batch --no-tty --dearmor -o /etc/apt/keyrings/docker.gpg
        sudo chmod a+r /etc/apt/keyrings/docker.gpg
        CODENAME=$(lsb_release -cs)
        if [ "$CODENAME" = "bookworm" ]; then
          CODENAME="bullseye"
        fi
        echo \
          "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") \
          $CODENAME stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt-get update -y
        echo "Docker repository configured"
      else
        echo "Docker GPG key already exists, skipping repository setup"
      fi

      sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin pgloader

      # Install Loki logging plugin
      if ! sudo docker plugin ls | grep -q "loki"; then
        sudo docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
      fi

      echo "Checking ssh keys"
      if ! grep "${MONITOR_PUB_KEY}" /root/.ssh/authorized_keys; then 
        echo "Adding pub key"
        echo "${MONITOR_PUB_KEY}" >> /root/.ssh/authorized_keys 
      else 
        echo "Pub key already added to system" 
      fi

      # --- SWARM SETUP (MANAGER) ---
      echo "Checking Docker Swarm status..."
      SWARM_IP=$(hostname -I | awk '{print $1}')

      if ! sudo docker info | grep "Swarm: active"; then
        echo "Swarm not active. Initializing as manager..."
        sudo docker swarm init --advertise-addr $SWARM_IP
      elif sudo docker info | grep -q "Is Manager: false"; then
        echo "Node is a worker in another swarm. Leaving and reinitializing as manager..."
        sudo docker swarm leave --force
        sudo docker swarm init --advertise-addr $SWARM_IP
      else
        echo "Already an active swarm manager, skipping init"
      fi

      

      # Label this node as the app node for placement constraints
      sudo docker node update --label-add role=app $(sudo docker node ls --format '{{.ID}}' --filter role=manager) || true

      # Save the worker join token and manager IP to a shared location
      # so the monitoring node can join the swarm
      WORKER_TOKEN=$(sudo docker swarm join-token -q worker)

      # Copy compose file
      mkdir -p /home/vagrant
      if [ -f /vagrant/$COMPOSE_FILE ]; then
        cp /vagrant/$COMPOSE_FILE /home/vagrant/
      elif [ -f /tmp/$COMPOSE_FILE ]; then
        cp /tmp/$COMPOSE_FILE /home/vagrant/
      else
        echo "ERROR: Could not find $COMPOSE_FILE in /vagrant or /tmp"
        exit 1
      fi

      # Prepare the volume for PostgreSQL
      if [ -b "$VOLUME_DEVICE" ]; then
        echo "Setting up volume for PostgreSQL..."
        sudo mkdir -p "$VOLUME_MOUNT"
        if ! mountpoint -q "$VOLUME_MOUNT"; then
          sudo mount "$VOLUME_DEVICE" "$VOLUME_MOUNT"
        fi
        if ! grep -q "$VOLUME_DEVICE" /etc/fstab; then
          echo "$VOLUME_DEVICE $VOLUME_MOUNT ext4 defaults,nofail 0 2" | sudo tee -a /etc/fstab
        fi
        sudo chown -R 999:999 "$VOLUME_MOUNT"
        sudo docker volume create \
          --driver local \
          --opt type=none \
          --opt device=$VOLUME_MOUNT \
          --opt o=bind \
          postgres_data || true
      else
        echo "WARNING: Volume device $VOLUME_DEVICE not found. Using local volume."
      fi

      envsubst < /home/vagrant/$COMPOSE_FILE > /home/vagrant/docker-compose.processed.yml

      cd /home/vagrant
      if sudo docker stack ls | grep -q "$STACK_NAME"; then
        echo "Stack $STACK_NAME already exists. Performing rolling update..."
        sudo docker pull $DOCKER_IMAGE
        sudo docker stack deploy \
          --compose-file /home/vagrant/docker-compose.processed.yml \
          --with-registry-auth \
          --prune \
          $STACK_NAME
        echo "Monitoring rolling update..."
        sleep 5
        sudo docker service ps ${STACK_NAME}_minitwit
      else
        echo "Stack $STACK_NAME not found. Deploying for first time..."
        sudo docker pull $DOCKER_IMAGE
        sudo docker pull postgres:15-alpine
        echo "Deploying Minitwit stack to Swarm..."
        sudo docker stack deploy -c /home/vagrant/docker-compose.processed.yml $STACK_NAME
        echo "Waiting for services to start..."
        sleep 15
        echo "PostgreSQL container logs:"
        sudo docker service logs ${STACK_NAME}_postgres --tail 20
      fi

      echo "Stack services:"
      sudo docker stack services $STACK_NAME
      echo "Stack ps:"
      sudo docker stack ps $STACK_NAME

      IP=$(hostname -I | awk '{print $1}')
      echo "===================================="
      echo "Minitwit deployed as Docker Swarm stack!"
      echo "===================================="
      echo "Minitwit app: http://$IP:8080"
      echo ""
      echo "Swarm commands:"
      echo "  docker stack services $STACK_NAME"
      echo "  docker stack ps $STACK_NAME"
      echo "  docker service logs ${STACK_NAME}_minitwit"
      echo "  docker service logs ${STACK_NAME}_postgres"
SHELL
  end

  config.vm.define "monitoring" do |monitor|
    monitor.vm.hostname = "monitoring"

    monitor.vm.provider :utm do |u, override|
      config.vm.synced_folder "./db", "/db" , owner: "root", group: "root"
      override.vm.box = "utm/bookworm"
      u.memory = 2048
      u.cpus = 2
    end

    monitor.vm.provider :libvirt do |lv, override|
      override.vm.box = "generic/ubuntu2204"
      lv.memory = 2048
      lv.cpus = 2
      lv.driver = "kvm"
      lv.default_prefix = ""
    end

    # DigitalOcean (Cloud)
    monitor.vm.provider :digital_ocean do |provider, override|
      override.vm.box = "digital_ocean"
      override.vm.box_url = "https://github.com/devopsgroup-io/vagrant-digitalocean/raw/master/box/digital_ocean.box"
      provider.token = ENV["DIGITAL_OCEAN_TOKEN"]
      provider.ssh_key_name = ENV["SSH_KEY_NAME"]
      override.ssh.private_key_path = '~/.ssh/devops_rsa'
      provider.image = "ubuntu-22-04-x64"
      provider.region = "fra1"
      provider.size = "s-1vcpu-1gb"
    end

    monitor.vm.network "forwarded_port", guest: 3000, host: 3000   # Grafana
    monitor.vm.network "forwarded_port", guest: 9090, host: 9090   # Prometheus
    monitor.vm.network "forwarded_port", guest: 3100, host: 3100   # Loki
    monitor.vm.provision "file", source: "./docker-compose-monitoring.yml", destination: "./docker-compose.yml"
    monitor.vm.provision "file", source: "./prometheus/prometheus_prod.yml", destination: "./prometheus/prometheus_prod.yml"
    monitor.vm.provision "file", source: "./loki/loki-config.yml", destination: "./loki/loki-config.yml"
    monitor.vm.provision "file", source: "./grafana", destination: "./grafana"
    monitor.vm.provision "file", source: "~/.ssh/id_monitor", destination: "/root/.ssh/id_monitor"

    monitor.vm.provision "shell", env: {
      "SWARM_MANAGER_IP" => ENV['SWARM_MANAGER_IP'],
    }, inline: <<-SHELL
      set -euo pipefail

      if [ -z "$SWARM_MANAGER_IP" ]; then
          echo "ERROR: $var is not set. Please set it in your host environment."
          exit 1
        fi

      sudo apt-get update -y
      sudo apt-get install -y ca-certificates curl gnupg lsb-release

      # Uninstall conflicting packages
      sudo apt remove --ignore-missing $(dpkg --get-selections docker.io docker-compose docker-compose-v2 docker-doc podman-docker containerd runc | cut -f1)

      if [ ! -f /etc/apt/keyrings/docker.gpg ]; then
        echo "Setting up Docker's GPG key and repository (first time only)..."
        sudo install -m 0755 -d /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg | sudo gpg --batch --no-tty --dearmor -o /etc/apt/keyrings/docker.gpg
        sudo chmod a+r /etc/apt/keyrings/docker.gpg
        CODENAME=$(lsb_release -cs)
        if [ "$CODENAME" = "bookworm" ]; then
          CODENAME="bullseye"
        fi
        echo \
          "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") \
          $CODENAME stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt-get update -y
        echo "Docker repository configured"
      else
        echo "Docker GPG key already exists, skipping repository setup"
      fi

      sudo apt-get update
      sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

      # --- SWARM SETUP (WORKER) ---
      eval "$(ssh-agent -s)"
      ssh-add /root/.ssh/id_monitor
      # Read the join token written by the minitwit (manager) node
      SWARM_WORKER_TOKEN=$(ssh root@$SWARM_MANAGER_IP "docker swarm join-token worker -q")
      if [ -n "$SWARM_WORKER_TOKEN" ] && [ -n "$SWARM_MANAGER_IP" ]; then
        if ! sudo docker info | grep "Swarm: active"; then
          echo "Joining swarm as worker..."
          sudo docker swarm join --token $SWARM_WORKER_TOKEN $SWARM_MANAGER_IP:2377
        else
          echo "Already part of a swarm"
        fi
      else
        echo "WARNING: SWARM_WORKER_TOKEN or SWARM_MANAGER_IP not set — minitwit node must be provisioned first."
        echo "Falling back to standalone docker compose..."
      fi

      # Label this node as the monitoring node for placement constraints
      # This must be run from the manager, so we SSH to it
      echo "Note: label 'role=monitoring' must be applied from the manager node:"
      MONITORING_NODE_ID=$(sudo docker info --format '{{.Swarm.NodeID}}')
      echo "Applying role=monitoring label to this node (ID: $MONITORING_NODE_ID)..."
      ssh root@$SWARM_MANAGER_IP "docker node update --label-add role=monitoring $MONITORING_NODE_ID"
      echo "Label applied successfully"
      

      mkdir -p ./prometheus_data
      sudo chown -R 65534:65534 ./prometheus_data
      mkdir -p ./grafana_data
      sudo chown -R 472:472 ./grafana_data
      sudo docker compose --profile prod down 2>/dev/null || true
      sudo docker compose --profile prod up -d

      IP=$(hostname -I | awk '{print $1}')
      echo "===================================="
      echo "Monitor deployed and running!"
      echo "===================================="
      echo "Access Prometheus at: http://$IP:9090"
      echo "Access Grafana at: http://$IP:3000"
      echo "Access Loki at: http://$IP:3100"
    SHELL
  end
end
# -*- mode: ruby -*-
# vi: set ft=ruby :

# Load .env into process environment so `vagrant up/provision` works without
# manually exporting variables. Existing exported vars take precedence.
env_file = File.expand_path('.env', __dir__)
if File.file?(env_file)
  File.readlines(env_file, chomp: true).each do |line|
    line = line.strip
    next if line.empty? || line.start_with?('#')

    line = line.sub(/^export\s+/, '')
    key, value = line.split('=', 2)
    next if key.nil? || value.nil?

    key = key.strip
    value = value.strip
    if (value.start_with?('"') && value.end_with?('"')) || (value.start_with?("'") && value.end_with?("'"))
      value = value[1..-2]
    end

    ENV[key] = value if ENV[key].nil? || ENV[key].empty?
  end
end

Vagrant.configure("2") do |config|

  config.vm.define "minitwit" do |server|
    server.vm.hostname = "minitwit"

    server.vm.provider :utm do |u, override|
      config.vm.synced_folder "./db", "/db" , owner: "root", group: "root"
      override.vm.box = "utm/bookworm"
      u.memory = 2048
      u.cpus = 2
      # UTM: upload to a guaranteed writable temp path
      override.vm.provision "file", source: "./docker-compose.yml", destination: "/tmp/docker-compose.yml"
      # UTM only: upload .env because shared folders are limited.
      if File.file?(env_file)
        override.vm.provision "file", source: env_file, destination: "/home/vagrant/.env"
      else
        warn "WARNING: .env was not found at #{env_file}. UTM guest will not receive /home/vagrant/.env"
      end
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
      override.ssh.private_key_path = '~/.ssh/' + ENV["SSH_KEY_NAME"]
      provider.image = "ubuntu-22-04-x64"
      provider.region = "fra1"
      provider.size = "s-1vcpu-1gb"
      provider.vm.provision "file", source: "./docker-compose.yml", destination: "/vagrant/docker-compose.yml"
    end

    # Local port forwarding (ignored by DO)
    server.vm.network "forwarded_port", guest: 8080, host: 8080

    # Copy nginx config directory from repo into VM (stage in /tmp, install in shell)
    server.vm.provision "file", source: "./nginx", destination: "/tmp/nginx"

    # Provisioning
    server.vm.provision "shell", env: {
      "USERNAME" => ENV['DOCKER_USERNAME'] || ENV['USERNAME'] || "flakiator",
      # Database credentials - NO DEFAULTS (must be set in host environment)
      "CONTAINER_NAME_PREFIX" => ENV['CONTAINER_NAME_PREFIX'],
      "POSTGRES_USER" => ENV['POSTGRES_USER'],
      "POSTGRES_PASSWORD" => ENV['POSTGRES_PASSWORD'],
      "POSTGRES_DB" => ENV['POSTGRES_DB'],
      "POSTGRES_HOST" => ENV['POSTGRES_HOST'],
      "POSTGRES_PORT" => ENV['POSTGRES_PORT'],
      "DB_SSL_MODE" => ENV['DB_SSL_MODE'],
      "VOLUME_MOUNT" => ENV['VOLUME_MOUNT'] || "/mnt/pgdata",
      "MONITOR_IP" => ENV['MONITOR_IP'],
      # NOIP ENVs
      "NOIP_USERNAME" => ENV['NOIP_USERNAME'],
      "NOIP_PASSWORD" => ENV['NOIP_PASSWORD'],
      "NOIP_HOSTNAMES" => ENV['NOIP_HOSTNAMES'],
      "CERTBOT_EMAIL" => ENV['CERTBOT_EMAIL'],
      # Resource limits - WITH DEFAULTS
      "POSTGRES_CPU_LIMIT" => ENV['POSTGRES_CPU_LIMIT'] || "0.2",
      "POSTGRES_MEM_LIMIT" => ENV['POSTGRES_MEM_LIMIT'] || "300M",
      "APP_REPLICAS" => ENV['APP_REPLICAS'] || "2",
      "APP_CPU_LIMIT" => ENV['APP_CPU_LIMIT'] || "0.2",
      "APP_MEM_LIMIT" => ENV['APP_MEM_LIMIT'] || "256M"
      }, inline: <<-SHELL
      set -euo pipefail

      VOLUME_DEVICE="/dev/sda"
      ARCH="$(uname -m)"
      if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        IMAGE_NAME="minitwitutmimage"
      else
        IMAGE_NAME="minitwitimage"
      fi
      DOCKER_IMAGE="$USERNAME/$IMAGE_NAME"
      export DOCKER_USERNAME="${DOCKER_USERNAME:-$USERNAME}"
      COMPOSE_FILE="docker-compose.yml"
      STACK_NAME="minitwit"


      # Check that env vars that are not default valued are actually set
      REQUIRED_VARS="POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_HOST MONITOR_IP"
      for var in $REQUIRED_VARS; do
        if [ -z "${!var}" ]; then
          echo "ERROR: $var is not set. Please set it in .env or your host environment."
          exit 1
        fi
      done

      # Docker installation
      # sudo apt-get remove -y docker.io docker-compose docker-compose-v2 docker-doc podman-docker containerd runc 2>/dev/null || true

      sudo apt-get update -y
      sudo apt-get install -y ca-certificates curl gnupg lsb-release

      # Only add GPG key and repository if not already present
      if [ ! -f /etc/apt/keyrings/docker.gpg ]; then
        echo "Setting up Docker's GPG key and repository (first time only)..."
        
        # Create keyrings directory
        sudo install -m 0755 -d /etc/apt/keyrings
        
        # Add Docker's GPG key
        curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg | sudo gpg --batch --no-tty --dearmor -o /etc/apt/keyrings/docker.gpg
        sudo chmod a+r /etc/apt/keyrings/docker.gpg

        # Add Docker repository
        CODENAME=$(lsb_release -cs)
        if [ "$CODENAME" = "bookworm" ]; then
          CODENAME="bullseye"
        fi

        echo \
          "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") $CODENAME stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

        sudo apt-get update -y
        

        echo "Docker repository configured"
      else
        echo "Docker GPG key already exists, skipping repository setup"
      fi

      sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin pgloader

      # Install loki plugin if possible (non-fatal on platforms where plugin is unavailable/broken)
      if ! sudo docker plugin ls --format '{{.Name}}' | grep -q '^loki$'; then
        if ! sudo docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions; then
          echo "WARNING: Loki docker logging plugin could not be installed. Falling back to default json-file logging."
        fi
      fi

      # Ensure plugin is enabled when present. If this fails, deployment will continue without loki driver.
      if sudo docker plugin inspect loki >/dev/null 2>&1; then
        sudo docker plugin enable loki >/dev/null 2>&1 || true
      fi
  
      echo "Checking Docker Swarm status..."
      if ! sudo docker info | grep "Swarm: active"; then
        echo "Swarm not active. Initializing..."
        # Get the main IP address (not localhost)
        SWARM_IP=$(hostname -I | awk '{print $1}')
        sudo docker swarm init --advertise-addr $SWARM_IP
      else
        echo "Swarm is active"
        # Check if this node is a manager
        if ! sudo docker node ls 2>/dev/null | grep -q "Leader"; then
          echo "This node is not a manager. Attempting to make it a manager..."
          # If it's a worker, we need to promote it (requires manager access)
          # This might fail if we don't have manager access, so we'll reinit if needed
          sudo docker swarm leave --force
          sudo docker swarm init --advertise-addr $(hostname -I | awk '{print $1}')
        fi
      fi

      # Verify swarm status
      sudo docker node ls || {
        echo "ERROR: Still not able to access swarm. Reinitializing..."
        sudo docker swarm leave --force 2>/dev/null || true
        sudo docker swarm init --advertise-addr $(hostname -I | awk '{print $1}')
      }
      # Copy compose file
      mkdir -p /home/vagrant
      if [ -f /vagrant/$COMPOSE_FILE ]; then
        # DO/VirtualBox/libvirt path (existing behavior)
        cp /vagrant/$COMPOSE_FILE /home/vagrant/
      elif [ -f /tmp/$COMPOSE_FILE ]; then
        # UTM fallback path
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
        
        # Set proper permissions for Docker volume
        sudo chown -R 999:999 "$VOLUME_MOUNT" 
        
        # Create a docker volume that uses this mount
        sudo docker volume create \
          --driver local \
          --opt type=none \
          --opt device=$VOLUME_MOUNT \
          --opt o=bind \
          postgres_data || true
      else
        echo "WARNING: Volume device $VOLUME_DEVICE not found. Using local volume."
      fi

      #  - - - - - - - Setup TLS - - - - - - - 
      sudo apt-get install -y ufw nginx
      sudo systemctl enable --now nginx

      # UFW enable is interactive unless --force is used.
      if ! sudo ufw status | grep -q "Status: active"; then
        sudo ufw --force enable
      else
        echo "UFW is already active"
      fi
      sudo ufw allow 'Nginx HTTP'
      sudo ufw allow ssh

      # No-IP dynamic DNS client (idempotent)
      NOIP_IMAGE="ghcr.io/noipcom/noip-duc:latest"
      NOIP_CONTAINER="noip-duc"
      NOIP_ENV_FILE="/home/vagrant/.env"
      if [ ! -f "$NOIP_ENV_FILE" ] && [ -f ".env" ]; then
        NOIP_ENV_FILE=".env"
      fi

      if [ -z "${NOIP_USERNAME:-}" ] || [ -z "${NOIP_PASSWORD:-}" ] || [ -z "${NOIP_HOSTNAMES:-}" ]; then
        echo "WARNING: NOIP_* variables are not fully set. Skipping noip-duc container startup."
      elif [ ! -f "$NOIP_ENV_FILE" ]; then
        echo "WARNING: $NOIP_ENV_FILE not found. Skipping noip-duc container startup."
      else
        sudo docker pull "$NOIP_IMAGE"
        if sudo docker ps -a --format '{{.Names}}' | grep -qx "$NOIP_CONTAINER"; then
          if [ "$(sudo docker inspect -f '{{.State.Running}}' "$NOIP_CONTAINER")" = "true" ]; then
            echo "$NOIP_CONTAINER is already running"
          else
            echo "Starting existing $NOIP_CONTAINER container"
            sudo docker start "$NOIP_CONTAINER"
          fi
        else
          sudo docker run -d \
            --env-file "$NOIP_ENV_FILE" \
            --restart unless-stopped \
            --name "$NOIP_CONTAINER" \
            "$NOIP_IMAGE"
        fi
      fi
      
      # Setup nginx config
      if [ -f /tmp/nginx/sites-available/minitwit.conf ]; then
        sudo install -m 0644 /tmp/nginx/sites-available/minitwit.conf /etc/nginx/sites-available/minitwit.conf
      elif [ -f /vagrant/nginx/sites-available/minitwit.conf ]; then
        sudo install -m 0644 /vagrant/nginx/sites-available/minitwit.conf /etc/nginx/sites-available/minitwit.conf
      else
        echo "WARNING: minitwit nginx site config not found in /tmp/nginx or /vagrant/nginx"
      fi

      if [ -f /etc/nginx/sites-available/minitwit.conf ]; then
        sudo ln -sfn /etc/nginx/sites-available/minitwit.conf /etc/nginx/sites-enabled/minitwit.conf
        sudo rm -f /etc/nginx/sites-enabled/default
        sudo nginx -t
        sudo systemctl reload nginx
      fi

      # Install certbot for TLS certificate management
      if ! command -v snap >/dev/null 2>&1; then
        sudo apt-get install -y snapd
      fi
      sudo systemctl enable --now snapd.socket
      if ! sudo snap list core >/dev/null 2>&1; then
        sudo snap install core
      fi
      sudo snap refresh core
      if ! sudo snap list certbot >/dev/null 2>&1; then
        sudo snap install --classic certbot
      fi

      # Create/update symlink for certbot
      sudo ln -sfn /snap/bin/certbot /usr/bin/certbot

      # Allow necessary ports through the firewall
      sudo ufw allow 'Nginx Full'
      if sudo ufw status | grep -q "Nginx HTTP"; then
        sudo ufw delete allow 'Nginx HTTP'
      fi
      sudo ufw allow ssh

      # Obtain/renew TLS certificate without interactive prompts
      if [ -n "${CERTBOT_EMAIL:-}" ] && [ -n "${NOIP_HOSTNAMES:-}" ]; then
        sudo certbot --nginx \
          --non-interactive \
          --agree-tos \
          --email "$CERTBOT_EMAIL" \
          --keep-until-expiring \
          --redirect \
          -d "$NOIP_HOSTNAMES"
      else
        echo "WARNING: CERTBOT_EMAIL or NOIP_HOSTNAMES is missing. Skipping certbot."
      fi
      
      # - - - - - - - End of TLS Setup - - - - - - -

      # Create a processed compose file with variables substituted
      envsubst < /home/vagrant/$COMPOSE_FILE > /home/vagrant/docker-compose.processed.yml

      # Force app image to match architecture-specific image selection.
      sed -i "s#/minitwitimage:latest#/$IMAGE_NAME:latest#g" /home/vagrant/docker-compose.processed.yml

      # Some platforms (notably UTM/arm64) may have a present but unusable loki plugin.
      # If loki is not enabled, remove only minitwit's logging block so stack deploy can proceed.
      LOKI_PLUGIN_ENABLED="$(sudo docker plugin inspect loki --format '{{.Enabled}}' 2>/dev/null || echo false)"
      if [ "$LOKI_PLUGIN_ENABLED" != "true" ]; then
        echo "Loki plugin is not enabled/healthy. Removing loki logging config from minitwit service for this deployment."
        awk '
          /^  minitwit:/ {in_minitwit=1}
          /^  [a-zA-Z0-9_-]+:/ && $0 !~ /^  minitwit:/ {in_minitwit=0}
          in_minitwit && /^    logging:/ {skip_logging=1; next}
          skip_logging && /^    [a-zA-Z0-9_-]+:/ {skip_logging=0}
          skip_logging {next}
          {print}
        ' /home/vagrant/docker-compose.processed.yml > /home/vagrant/docker-compose.processed.tmp.yml
        mv /home/vagrant/docker-compose.processed.tmp.yml /home/vagrant/docker-compose.processed.yml
      fi

      # Check if stack exists
      cd /home/vagrant
      if sudo docker stack ls | grep -q "$STACK_NAME"; then
        echo "Stack $STACK_NAME already exists. Performing rolling update..."
        
        # Pull the latest image
        echo "Pulling latest image: $DOCKER_IMAGE"
        sudo docker pull $DOCKER_IMAGE 
        
        
        # Update the service with rolling update
        echo "Starting rolling update of minitwit service..."
        # sudo docker service update \
        #   --image $DOCKER_IMAGE \
        #   --force \
        #   --update-parallelism 1 \
        #   --update-delay 10s \
        #   --update-order start-first \
        #   --update-failure-action rollback \
        #   ${STACK_NAME}_minitwit

        sudo docker stack deploy \
          --compose-file /home/vagrant/docker-compose.processed.yml \
          --with-registry-auth \
          --prune \
          $STACK_NAME
        
        # Monitor the update
        echo "Monitoring rolling update..."
        sleep 5
        sudo docker service ps ${STACK_NAME}_minitwit
        
        echo "Rolling update initiated! You can monitor with:"
        echo "  docker service ps ${STACK_NAME}_minitwit"
        echo "  docker service logs ${STACK_NAME}_minitwit -f"
        
      else
        echo "Stack $STACK_NAME not found. Deploying for first time..."
        
        # Pull images
        echo "Pulling latest images..."
        sudo docker pull $DOCKER_IMAGE
        sudo docker pull postgres:15-alpine
        
        # Deploy the stack using the processed compose file
        echo "Deploying Minitwit stack to Swarm..."
        sudo docker stack deploy -c /home/vagrant/docker-compose.processed.yml $STACK_NAME
        
        # Wait for services to start
        echo "Waiting for services to start..."
        sleep 15
        
        # Show PostgreSQL logs to verify it started correctly
        echo "PostgreSQL container logs:"
        sudo docker service logs ${STACK_NAME}_postgres --tail 20
      fi

      # DATABASE MIGRATION COMMAND GOES HERE


      # Check final status
      echo "Stack services:"
      sudo docker stack services $STACK_NAME

      echo "Stack ps:"
      sudo docker stack ps $STACK_NAME

      # Get the IP for accessing the services
      IP=$(hostname -I | awk '{print $1}')

      echo "===================================="
      echo "Minitwit deployed as Docker Swarm stack!"
      echo "===================================="
      echo "Minitwit app: http://$IP:8080 (available on all swarm nodes)"
      echo "PostgreSQL is running as a container in the stack"
      echo ""
      echo "Stack name: $STACK_NAME"
      echo "Services:"
      sudo docker stack services $STACK_NAME --format "table {{.Name}}\t{{.Replicas}}\t{{.Ports}}"
      echo ""
      echo "Rolling update config (from compose file):"
      echo "  - Parallelism: 1 container at a time"
      echo "  - Delay: 10s between updates"
      echo "  - Order: start-first (new starts before old stops)"
      echo "  - Failure action: rollback"
      echo ""
      echo "Swarm commands:"
      echo "  docker stack services $STACK_NAME           # List services"
      echo "  docker stack ps $STACK_NAME                 # List tasks"
      echo "  docker service logs ${STACK_NAME}_minitwit  # View app logs"
      echo "  docker service logs ${STACK_NAME}_postgres  # View postgres logs"
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
    monitor.vm.provision "shell", inline: <<-SHELL
      sudo apt-get update -y
      sudo apt-get install -y ca-certificates curl gnupg lsb-release
      
      # Uninstall conflicting packages
      sudo apt remove --ignore-missing $(dpkg --get-selections docker.io docker-compose docker-compose-v2 docker-doc podman-docker containerd runc | cut -f1)
      
      # Only add GPG key and repository if not already present
      if [ ! -f /etc/apt/keyrings/docker.gpg ]; then
        echo "Setting up Docker's GPG key and repository (first time only)..."
        
        # Create keyrings directory
        sudo install -m 0755 -d /etc/apt/keyrings
        
        # Add Docker's GPG key
        curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg | sudo gpg --batch --no-tty --dearmor -o /etc/apt/keyrings/docker.gpg
        sudo chmod a+r /etc/apt/keyrings/docker.gpg

        # Add Docker repository
        CODENAME=$(lsb_release -cs)
        if [ "$CODENAME" = "bookworm" ]; then
          CODENAME="bullseye"
        fi

        echo \
          "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") $CODENAME stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

        sudo apt-get update -y
        
        echo "Docker repository configured"
      else
        echo "Docker GPG key already exists, skipping repository setup"
      fi

      # Update package list again (for some reason it won't work if we don't)
      sudo apt-get update

      # Docker engine
      sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
      
      mkdir -p ./prometheus_data
      sudo chown -R 65534:65534 ./prometheus_data # Set correct user perms for prometheus
      mkdir -p ./grafana_data
      sudo chown -R 472:472 ./grafana_data # Set correct user perms for grafana
      sudo docker compose --profile prod down
      sudo docker compose --profile prod up -d
      
      echo "===================================="
      echo "Monitor deployed and running!"
      echo "===================================="

      IP=$(hostname -I | awk '{print $1}')
      echo "Access Prometheus at: http://$IP:9090"    
      echo "Access Grafana at: http://$IP:3000"
    SHELL
  end
end

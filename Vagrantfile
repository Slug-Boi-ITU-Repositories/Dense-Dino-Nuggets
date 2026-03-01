# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|

  config.vm.synced_folder ".", "/vagrant"

  config.vm.define "minitwit" do |server|
    server.vm.hostname = "minitwit"

    server.vm.provider :utm do |u, override|
      override.vm.box = "utm/bookworm"
      u.memory = 2048
      u.cpus = 2
    end

    server.vm.provider :virtualbox do |vb, override|
      override.vm.box = "ubuntu/jammy64"
      vb.memory = 2048
      vb.cpus = 2
    end


    # DigitalOcean (Cloud)
    server.vm.provider :digital_ocean do |provider, override|
      override.vm.box = "digital_ocean"
      override.vm.box_url = "https://github.com/devopsgroup-io/vagrant-digitalocean/raw/master/box/digital_ocean.box"
      override.vm.synced_folder ".", "/vagrant", type: "rsync"
      provider.token = ENV["DIGITAL_OCEAN_TOKEN"]
      provider.ssh_key_name = ENV["SSH_KEY_NAME"]
      override.ssh.private_key_path = '~/.ssh/devops_rsa'
      provider.image = "ubuntu-22-04-x64"
      provider.region = "fra1"
      provider.size = "s-1vcpu-1gb"
    end

    # Local port forwarding (ignored by DO)
    server.vm.network "forwarded_port", guest: 8080, host: 8080


    # Provisioning
    server.vm.provision "shell",env: {"USERNAME" => ENV['DOCKER_USERNAME']} ,inline: <<-SHELL
      sudo apt-get update -y
      sudo apt-get install -y ca-certificates curl gnupg lsb-release
      
      # Uninstall conflicting packages
      sudo apt remove $(dpkg --get-selections docker.io docker-compose docker-compose-v2 docker-doc podman-docker containerd runc | cut -f1)
      
      # 3. Add Docker's official GPG key
      sudo install -m 0755 -d /etc/apt/keyrings
      curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
      sudo chmod a+r /etc/apt/keyrings/docker.gpg

      # 4. Set up the Docker repository
      # This check was written by Gemini
      # NOTE: For Debian Bookworm, we must use the repository for Bullseye as Docker doesn't have a Bookworm repo yet.
      CODENAME=$(lsb_release -cs)
      if [ "$CODENAME" = "bookworm" ]; then
        CODENAME="bullseye"
      fi
      echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") \
        $CODENAME stable" | \
        sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

      # Update package list again (for some reason it won't work if we don't)
      sudo apt-get update

      # Docker engine
      sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

      # Stop and remove any existing container named minitwitimage
      IMAGE_NAME="minitwitimage"
      if [ "$(sudo docker ps -q -f name=$IMAGE_NAME)" ]; then
          sudo docker stop $IMAGE_NAME
      fi
      if [ "$(sudo docker ps -aq -f status=exited -f name=$IMAGE_NAME)" ]; then
          sudo docker rm $IMAGE_NAME
      fi

      # Set default image name if environment variable is not set or empty
      DOCKER_IMAGE=$USERNAME/$IMAGE_NAME

      # Pull the latest image and run the container
      sudo docker run -d --pull always --name $IMAGE_NAME -p 8080:8080 "$DOCKER_IMAGE"


      echo "===================================="
      echo "Minitwit deployed from $DOCKER_IMAGE!"
      echo "===================================="

      IP=$(hostname -I | awk '{print $1}')
      echo "Access at: http://$IP:8080"
    SHELL

  end
end
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
    server.vm.provision "shell", inline: <<-SHELL
      sudo apt-get update -y
      
      echo "=== Removing system Go (if any) ==="
      sudo apt-get remove -y golang-go golang-* 2>/dev/null || true
      sudo apt-get autoremove -y || true

      # CGO dependencies
      sudo apt-get install -y gcc libsqlite3-dev
      
      GO_VERSION="1.25.5"
  # Detect architecture of machine
       ARCH=$(uname -m)
  case $ARCH in
    x86_64)
      GO_ARCH="amd64"
      ;;
    aarch64)
      GO_ARCH="arm64"
      ;;
    *)
      echo "Unsupported architecture: $ARCH"
      exit 1
      ;;
  esac

      #sudo apt-get install -y golang-go
      wget -q "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
      sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
      rm "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"

      # Ensure new Go is first in path
      echo 'export PATH=/usr/local/go/bin:$PATH' | sudo tee /etc/profile.d/go.sh
      source /etc/profile.d/go.sh

      # verify go version
      echo "=== Go environment ==="
      which go
      go version
      
      mkdir /home/vagrant
      cp -r /vagrant/* /home/vagrant/
      cd /home/vagrant

      export CGO_ENABLED=1   # explicitly enable CGO
      go mod tidy
      go build -o minitwit ./src
      echo "grab a cup of coffee this step takes a minute"
      nohup ./minitwit > app.log 2>&1 &

      echo "===================================="
      echo "Minitwit deployed!"
      echo "===================================="

      IP=$(hostname -I | awk '{print $1}')
      echo "Access at: http://$IP:8080"
    SHELL

  end
end
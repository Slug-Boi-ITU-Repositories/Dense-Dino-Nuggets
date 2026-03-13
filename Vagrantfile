# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|

  config.vm.define "minitwit" do |server|
    server.vm.hostname = "minitwit"
      server.vm.network "private_network", ip: "192.168.56.10"

    server.vm.provider :utm do |u, override|
      config.vm.synced_folder "./db", "/db" , owner: "root", group: "root"
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
      provider.token = ENV["DIGITAL_OCEAN_TOKEN"]
      provider.ssh_key_name = ENV["SSH_KEY_NAME"]
      override.ssh.private_key_path = '~/.ssh/devops_rsa'
      provider.image = "ubuntu-22-04-x64"
      provider.region = "fra1"
      provider.size = "s-1vcpu-1gb"
      provider.private_networking = true
    end

    # Local port forwarding (ignored by DO)
    server.vm.network "forwarded_port", guest: 8080, host: 8080

    # Provisioning
    server.vm.provision "shell",env: {"USERNAME" => ENV['DOCKER_USERNAME'], "DATABASE_URL" => ENV['DATABASE_URL']} , path: "vagrant_shell/provision-app.sh"

  end

#-----------------------------
#Postgres Database Server
#-----------------------------
config.vm.define "postgres" do |db|
db.vm.hostname = "postgres"
db.vm.network "private_network", ip: "192.168.56.11"
# UTM
db.vm.provider :utm do |u, override|
  override.vm.box = "utm/bookworm"
  u.memory = 1024
  u.cpus = 1
end

# VirtualBox
db.vm.provider :virtualbox do |vb, override|
  override.vm.box = "ubuntu/jammy64"
  vb.memory = 1024
  vb.cpus = 1
end

# DigitalOcean
db.vm.provider :digital_ocean do |provider, override|
  override.vm.box = "digital_ocean"
  override.vm.box_url = "https://github.com/devopsgroup-io/vagrant-digitalocean/raw/master/box/digital_ocean.box"
  provider.token = ENV["DIGITAL_OCEAN_TOKEN"]
  provider.ssh_key_name = ENV["SSH_KEY_NAME"]
  override.ssh.private_key_path = '~/.ssh/devops_rsa'
  provider.image = "ubuntu-22-04-x64"
  provider.region = "fra1"
  provider.size = "s-1vcpu-1gb"
  provider.private_networking = true
end

db.vm.network "forwarded_port", guest: 5432, host: 5432
db.vm.provision "shell", path: "vagrant_shell/provision-postgres.sh"

end

end
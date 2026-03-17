# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|

  # Bring the database VM up before the app VM so app provisioning can rely on it.
  config.vm.define "minitwit" do |server|
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
    docker_username = ENV.fetch('DOCKER_USERNAME') { abort "Set the DOCKER_USERNAME env var before running vagrant up" }
    server.vm.provision "shell",
      env: {
        "USERNAME"     => docker_username,
        "DATABASE_URL" => ENV.fetch('DATABASE_URL', '')
      },
      path: "vagrant_shell/provision-app.sh"

  end

end

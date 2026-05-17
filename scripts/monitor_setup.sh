#!/bin/bash
# Save stdout to fd3, redirect all stdout to stderr to allow for env var capture later
exec 3>&1 1>&2 
set -euo pipefail

if [ -z "$SWARM_MANAGER_IP" ]; then
    echo "ERROR: $var is not set. Please set it in your host environment."
    exit 1
fi


sudo apt-get update -y
sudo apt-get install -y ca-certificates curl gnupg lsb-release ufw

ufw allow 2377/tcp
ufw allow 7946
ufw allow 4789/udp

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


# Grab node label and ip for use in the manager VM
NODE_ID=$(sudo docker info --format '{{.Swarm.NodeID}}') 
IP=$(hostname -I | awk '{print $1}')
echo "$NODE_ID $IP" >&3



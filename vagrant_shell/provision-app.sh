#!/bin/bash
set -euo pipefail

VOLUME_MOUNT="/mnt/pgdata"
VOLUME_DEVICE="/dev/sda"
IMAGE_NAME="minitwitimage"
DOCKER_IMAGE="$USERNAME/$IMAGE_NAME"
DATABASE_URL="postgres://philip:admin@localhost:5432/minitwit?sslmode=disable"

sudo apt-get update -y
sudo apt-get install -y ca-certificates curl gnupg lsb-release postgresql postgresql-contrib

sudo systemctl enable postgresql
sudo systemctl start postgresql

read -r PG_VERSION PG_CLUSTER < <(sudo pg_lsclusters --no-header | awk 'NR == 1 {print $1, $2}')

if [[ -z "$PG_VERSION" || -z "$PG_CLUSTER" ]]; then
  echo "Could not determine PostgreSQL cluster." >&2
  exit 1
fi

sudo systemctl stop postgresql

if [ -b "$VOLUME_DEVICE" ]; then
  sudo mkdir -p "$VOLUME_MOUNT"
  if ! mountpoint -q "$VOLUME_MOUNT"; then
    mount "$VOLUME_DEVICE" "$VOLUME_MOUNT"
  fi
  if ! grep -q "$VOLUME_DEVICE" /etc/fstab; then
    echo "$VOLUME_DEVICE $VOLUME_MOUNT ext4 defaults,nofail 0 2" | sudo tee -a /etc/fstab
  fi
else
  sudo mkdir -p "$VOLUME_MOUNT"
fi

if [ -z "$(ls -A $VOLUME_MOUNT)" ]; then
  sudo -u postgres /usr/lib/postgresql/"$PG_VERSION"/bin/initdb -D "$VOLUME_MOUNT"
fi

sudo chown -R postgres:postgres "$VOLUME_MOUNT"
sudo pg_conftool "$PG_VERSION" "$PG_CLUSTER" set data_directory "$VOLUME_MOUNT"

PG_HBA="/etc/postgresql/$PG_VERSION/$PG_CLUSTER/pg_hba.conf"
ACCESS_RULE_IPV4="host all all 0.0.0.0/0 scram-sha-256"
ACCESS_RULE_IPV6="host all all ::/0 scram-sha-256"

if ! sudo grep -qxF "$ACCESS_RULE_IPV4" "$PG_HBA"; then
  echo "$ACCESS_RULE_IPV4" | sudo tee -a "$PG_HBA" >/dev/null
fi

if ! sudo grep -qxF "$ACCESS_RULE_IPV6" "$PG_HBA"; then
  echo "$ACCESS_RULE_IPV6" | sudo tee -a "$PG_HBA" >/dev/null
fi

sudo systemctl start postgresql

sudo -u postgres psql <<'EOF'
DO
$do$
BEGIN
  IF NOT EXISTS (
    SELECT FROM pg_catalog.pg_roles WHERE rolname = 'philip'
  ) THEN
    CREATE ROLE philip LOGIN PASSWORD 'admin';
  END IF;
END
$do$;

SELECT 'CREATE DATABASE minitwit OWNER philip'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'minitwit')\gexec
EOF

if ! pg_isready -h 127.0.0.1 -p 5432 -U philip -d minitwit >/dev/null 2>&1; then
  echo "PostgreSQL did not become ready." >&2
  exit 1
fi

sudo apt-get remove -y docker.io docker-compose docker-compose-v2 docker-doc podman-docker containerd runc 2>/dev/null || true

sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

CODENAME=$(lsb_release -cs)
if [ "$CODENAME" = "bookworm" ]; then
  CODENAME="bullseye"
fi

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") \
  $CODENAME stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

if [ "$(sudo docker ps -q -f name=$IMAGE_NAME)" ]; then
  sudo docker stop $IMAGE_NAME
fi
if [ "$(sudo docker ps -aq -f name=$IMAGE_NAME)" ]; then
  sudo docker rm $IMAGE_NAME
fi

sudo docker run -d --pull always \
  --name $IMAGE_NAME \
  --network host \
  -e DATABASE_URL="$DATABASE_URL" \
  "$DOCKER_IMAGE"

echo "===================================="
echo "Minitwit deployed from $DOCKER_IMAGE!"
echo "===================================="

IP=$(hostname -I | awk '{print $1}')
echo "Access at: http://$IP:8080"
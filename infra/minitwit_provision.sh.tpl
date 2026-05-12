#!/bin/bash
export POSTGRES_USER="${postgres_user}"
export POSTGRES_PASSWORD="${postgres_password}"
export POSTGRES_DB="${postgres_db}"
export VOLUME_MOUNT="${volume_mount}"
export POSTGRES_HOST="${postgres_host}"
export POSTGRES_PORT="${postgres_port}"
export DOCKER_USERNAME="${docker_username}"
export MONITOR_IP="${monitor_ip}"
export CONTAINER_NAME_PREFIX="${container_name_prefix}"
export DB_SSL_MODE="${db_ssl_mode}"

set -euo pipefail

VOLUME_DEVICE="/dev/sda"
IMAGE_NAME="minitwitimage"
# For UTM use the dockerhub image for arm64
#IMAGE_NAME="minitwitutmimage"
DOCKER_IMAGE="$DOCKER_USERNAME/$IMAGE_NAME"
COMPOSE_FILE="docker-compose.yml"
STACK_NAME="minitwit"


# Check that env vars that are not default valued are actually set
REQUIRED_VARS="POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_HOST"
for var in $REQUIRED_VARS; do
if [ -z "$${!var}" ]; then
    echo "ERROR: $var is not set. Please set it in your host environment."
    exit 1
fi
done

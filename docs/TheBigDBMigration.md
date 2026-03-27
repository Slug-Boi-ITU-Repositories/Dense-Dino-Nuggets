# The DB Migration Saga

## Introduction

This document outlines everything that we did to migrate our database and will document the process for those curious or if you need to do the same yourself.

This was orchistrated by multiple people below are the participants of the migration on the night (in case you need to contact us in regards to anything).  
Participants:

- @Slug-Boi
- @Flakiator
- @August-Brandt

## The setup

To begin the migration process we first needed to completely rewrite the vagrant shell script to use our new docker-compose setup which connects a swarm of our application to a postgres database container. This required the machine hosted on Digital Ocean to have all the required docker packages as well as the code to check for swarm setup and rolling updates of the application containers. This was a really big rewrite that was tested at least 10+ times on a fresh DO droplet to make sure it actually behaved in an idiomatic way.

## Setting up the real container

The vagrant script would have most likely failed or at worst killed our entire app if ran unsupervised against the container, instead we ran the script piecewise to first setup all the packages and keys on the container and then lastly we preped a shell script on the machine that would kill the old app and setup the swarm to apply the schema to the new postgres DB and lastly copy all the data over from our SQLITE3 DB into the new DB.

### The setup script

Below is the script that was ran initially to setup the container in preperation for the big move.

<details>

<summary>Setup scritp</summary>

```bash
      set -euo pipefail

      VOLUME_DEVICE="/dev/sda"
      IMAGE_NAME="minitwitimage"
      # For UTM use the dockerhub image for arm64
      #IMAGE_NAME="minitwitutmimage"
      DOCKER_IMAGE="$USERNAME/$IMAGE_NAME"
      COMPOSE_FILE="docker-compose.yml"
      STACK_NAME="minitwit"


      # Check that env vars that are not default valued are actually set
      REQUIRED_VARS="POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_HOST"
      for var in $REQUIRED_VARS; do
        if [ -z "${!var}" ]; then
          echo "ERROR: $var is not set. Please set it in your host environment."
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
          "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") \
          $CODENAME stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

        sudo apt-get update -y
        
        echo "Docker repository configured"
      else
        echo "Docker GPG key already exists, skipping repository setup"
      fi

      sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin pgloader

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

      # Create a processed compose file with variables substituted
      envsubst < /home/vagrant/$COMPOSE_FILE > /home/vagrant/docker-compose.processed.yml

```

</details>

## The big move

We decided to wait with the move till late in the day 21:00 as the app seemed slow in user interaction at this time. Once the time came we ran the rest of the provisioning script which can be found below.

<details>

<summary>Kill and switch script</summary>

```bash
      # Check if stack exists
      cd /home/vagrant
      if sudo docker stack ls | grep -q "$STACK_NAME"; then
        echo "Stack $STACK_NAME already exists. Performing rolling update..."
        
        # Pull the latest image
        echo "Pulling latest image: $DOCKER_IMAGE"
        sudo docker pull $DOCKER_IMAGE 
        
        
        # Update the service with rolling update
        echo "Starting rolling update of minitwit service..."
        sudo docker service update \
          --image $DOCKER_IMAGE \
          --force \
          --update-parallelism 1 \
          --update-delay 10s \
          --update-order start-first \
          --update-failure-action rollback \
          ${STACK_NAME}_minitwit
        
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
      cat > /tmp/pgloader.cmd <<-EOF
      LOAD DATABASE FROM sqlite:///db/minitwit.db
      INTO postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}

      WITH include drop, create tables, create indexes, reset sequences,
          disable triggers, batch rows = 10000, batch concurrency = 1

      CAST type string to text drop typemod,
          type datetime to timestamptz drop default drop not null using zero-dates-to-null,
          type date to date drop default drop not null using zero-dates-to-null,
          type boolean to boolean using tinyint-to-boolean;
EOF

      # Run pgloader with the command file
      pgloader /tmp/pgloader.cmd

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
```

</details>

The most interesting part of this is the actual DB migration part which can be found in the section below.

### Migration command from SQLITE3 to Postgres

Paste the command below into your vagrant shell script after the DB has been setup using postgres and is running in a docker container. This will migrate all the data from the SQLITE3 database into the new postgres DB. WARNING, this will override any data in the postgres DB do not do this on a non empty DB or you will lose data. Fun fact if you move the ending EOF inline with the rest of the script the command explodes.

```bash
      cat > /tmp/pgloader.cmd <<-EOF
      LOAD DATABASE FROM sqlite:///db/minitwit.db
      INTO postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}

      WITH include drop, create tables, create indexes, reset sequences,
          disable triggers, batch rows = 10000, batch concurrency = 1

      CAST type string to text drop typemod,
          type datetime to timestamptz drop default drop not null using zero-dates-to-null,
          type date to date drop default drop not null using zero-dates-to-null,
          type boolean to boolean using tinyint-to-boolean;
EOF

      # Run pgloader with the command file
      pgloader /tmp/pgloader.cmd
```

## Slow application

After the migration we monitored and tested the system manually and noticed that the frontend seemed extremly slow (2+ seconds response time on requests), but only on the real frontend not the API endpoints. Since we only had logging for the endpoints this was very hard to debug. We added logging output to the frontend and eventually figured out that it was the SQL queres that were extremely slow to respond. We eventually settled on it being an indexing problem and manually applied these indexing rules to our database.

```sql
CREATE UNIQUE INDEX CONCURRENTLY idx_user_username ON "user"(username);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_message_author ON message(author_id);
CREATE INDEX CONCURRENTLY idx_message_flagged_pubdate ON message(flagged, pub_date DESC);
CREATE INDEX CONCURRENTLY idx_message_author_pubdate  ON message(author_id, pub_date DESC);
CREATE UNIQUE INDEX CONCURRENTLY idx_follower_who_whom ON follower(who_id, whom_id);
```

We also noticed that one the queries we had generated with GORM was very slow, it turns out doing OR in GORM results in a full search on both queries on all rows... This is the worst possible scenario and no matter what would result in really bad running times. We settled on rewriting the command to instead do 2 GORM based queries to get what the results of the 2 branches of the OR statement was and afterwards utilized a Postgres specific command called `UNION ALL` which helps us union the 2 branches removing all duplicates. This new query is near instant as it now respects the indexing rules we've applied to the tables and no longer does a full search of all the rows twice. All this can be found in the DB query function in our go files located in the repository folder.

After all these changes the app seemed responsive again and we've started monitoring on a seperate droplet with Prometheus and Grafana.

## Metrics

We seem to have had a downtime period of less than 5 minutes (closer to 3 minutes) which was very good and from our very rough estimates we seem to have lost/dropped basically no register requests during this time which would in turn result in no continued errors of users trying to follow or tweet from non-existant users. This is a very optimal outcome and the group is very happy with the migration process.

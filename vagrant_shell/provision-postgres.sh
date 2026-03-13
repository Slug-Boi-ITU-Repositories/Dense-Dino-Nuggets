#!/bin/bash
set -e

sudo apt-get update -y
sudo apt-get install -y postgresql postgresql-contrib

# Enable postgres service
sudo systemctl enable postgresql
sudo systemctl start postgresql

# Allow external connections
PG_VERSION=$(ls /etc/postgresql)
POSTGRES_CONF="/etc/postgresql/$PG_VERSION/main/postgresql.conf"
PG_HBA="/etc/postgresql/$PG_VERSION/main/pg_hba.conf"

sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/g" $POSTGRES_CONF

grep -qxF "host all all 0.0.0.0/0 md5" $PG_HBA || echo "host all all 0.0.0.0/0 md5" | sudo tee -a $PG_HBA

sudo -u postgres psql <<'EOF'
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles WHERE rolname = 'minitwit'
   ) THEN
      CREATE ROLE minitwit LOGIN PASSWORD 'minitwitpassword';
   END IF;
END
$do$;

SELECT 'CREATE DATABASE minitwit OWNER minitwit'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'minitwit')\gexec
EOF

sudo systemctl restart postgresql

echo "PostgreSQL ready."
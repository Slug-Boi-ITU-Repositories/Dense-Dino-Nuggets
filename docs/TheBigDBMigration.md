# The DB Migration Saga

## Introduction

This document outlines everything that we did to migrate our database and will document the process for those curious or if you need to do the same yourself.

## The setup

To begin the migration process we first needed to completely rewrite the vagrant shell script to use our new docker-compose setup which connects a swarm of our application to a postgres database container. This required the machine hosted on Digital Ocean to have all the required docker packages as well as the code to check for swarm setup and rolling updates of the application containers. This was a really big rewrite that was tested at least 10+ times on a fresh DO droplet to make sure it actually behaved in an idiomatic way.

## Setting up the real container

The vagrant script would have most likely failed or at worst killed our entire app if ran unsupervised against the container, instead we ran the script piecewise to first setup all the packages and keys on the container and then lastly we preped a shell script on the machine that would kill the old app and setup the swarm to apply the schema to the new postgres DB and lastly copy all the data over from our SQLITE3 DB into the new DB.

### The setup script

Below is the script that was ran initially to setup the container in preperation for the big move.

```bash

```

## The big move

We decided to wait with the move till late in the day 21:00 as the app seemed slow in user interaction at this time. Once the time came we ran the rest of the provisioning script which can be found below.

```bash

```

The most interesting part of this is the actual DB migration part which can be found in the section below.

### Migration command from SQLITE3 to Postgres

Paste the command below into your vagrant shell script after the DB has been setup using postgres and is running in a docker container. This will migrate all the data from the SQLITE3 database into the new postgres DB. WARNING, this will override any data in the postgres DB do not do this on a non empty DB or you will lose data.

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

```

We also noticed that one the queries we had generated with GORM was very slow, it turns out doing OR in GORM results in a full search on both queries on all rows... This is the worst possible scenario and no matter what would result in really bad running times. We settled on rewriting the command to instead do 2 GORM based queries to get what the results of the 2 branches of the OR statement was and afterwards utilized a Postgres specific command called `UNION ALL` which helps us union the 2 branches removing all duplicates. This new query is near instant as it now respects the indexing rules we've applied to the tables and no longer does a full search of all the rows twice.

After all these changes the app seemed responsive again and we've started monitoring on a seperate droplet with Prometheus and Grafana.

## Metrics

We seem to have had a downtime period of less than 5 minutes (closer to 3 minutes) which was very good and from our very rough estimates we seem to have lost/dropped basically no register requests during this time which would in turn result in no continued errors of users trying to login to non-existant users. This is a very optimal outcome and the group is very happy with the migration process.

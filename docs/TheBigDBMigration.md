# The DB Migration Saga

This document outlines everything that we did to migrate our database and will document the process for those curious or if you need to do the same yourself.

## Migration command from SQLITE3 to Postgres

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
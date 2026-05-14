# Diary of Dense Dino Nuggets

## 08/05

We worked on replacing the vagrant provisioning with Opentofu (a fork of Terraform). We created an Ansible configuration to ensure idempotency and to have a less convoluted format for this essential step in our pipeline.

## 21/04

The big swarm manager convergance migration. There is a seperate markdown file named `SwarmMigration.md` that outlines exactly how this came to be and what our measured downtime was.

## 19/03

This was the big DB migration night, we ended up writing a long document outlining what happened during this move and how it was all orchistrated. This file can be found in this folder under the name of `TheBigDBMigration.md`.

## 13/03

We have set up spellcheck. We have chosen to use misspell. It seems simple and is made specifically for Go. We were also considering CSpell but it seems like it would add unnecessary complexity and overlaps with the features of sonarqube.

## 06/03

We set up tests and linter workflows on PRs. We set up Dagger for this.
We chose Dagger so we can test workflows locally and because it makes it easier to
migrate our repo ("somewhat" decoupled from github)

We are mounting our database to the container running our app.
(Because our database is a file it does not make sense to host it on a different container)

We were behind on documentation so we had a session where we worked on it.

## 27/02

We have started implementing GORM. We have decided to use GORM since it is one of the largest
ORM packages in golang. Apparently most forum users do not think ORM's are useful on golang based threads.
We went with it anyways since we are actually migrating away from SQLITE3 (and because it was a requirement)
[Link to medium article about why ORM bad](https://medium.com/@enverbisevac/you-dont-need-orm-in-go-9216fb74cdfd)

We have refactored the tests from Python to Golang.

We are now hosting our docker image on DockerHub

We made a Vagrant provisioning script

The simulation is now running. We fixed the ReadTimeout errors we got.

## 18/02

Use `openapi-generator generate -i swagger3.json -g go-server -o ./out` to generate apimodels for the webapp apimodels.
Imported the generated go module into the webapp.

Make cookies local for each user session (instead of global)

## 13/02

We have merged a lot of PRs and resolved a lot of issues,
regarding the refactoring of minitwit.

## 06/02

We have started Refactoring minitwit from Python to Golang.

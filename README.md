# Dense-Dino-Nuggets - Minitwit

This is a fork of a student project of a small twitter like application (this is lore). Currently this is a golang server application that can be used to boot up a small webpage where people can, create accounts, login and post chirps. There is a also logic for following and unfollowing other users so you can get your own custom timeline to look at.

## Installation

### Environment setup

The program requires some environment variables to be set. The minimum for running the application locally is this

```bash
DATABASE_URL="<psql-dsn-string>"
```

Full env setup for docker-compose file. Can also be deployed as a stack on a swarm

```bash
set -a
source .env
set +a
```

You need an env file file that includes these vars or you need to have them exported in your environment if you are using a .env file you can export them to your env by using the following command. The `JWT_KEY` is what is used to sign and validate the JWT. This key should be kept secret and should be a long, hard to guess key. All instances of the server need to have the same key otherwise validation of the JWT issued by a different server will fail.

```bash
# Database Configuration
POSTGRES_USER=minitwit_user
POSTGRES_PASSWORD=minitwit_password
POSTGRES_DB=minitwit_db
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
DB_SSL_MODE=disable

# Container Names
CONTAINER_NAME_PREFIX=minitwit

# Resource Limits
POSTGRES_CPU_LIMIT=0.5
POSTGRES_MEM_LIMIT=512M
APP_REPLICAS=3

# JWT signing key
JWT_KEY=<Signing key>
```

### Local Docker Compose workflow

The following setup is used for local development from the project root. Remember to setup a .env file as specified in the [Environment setup] section.
Load the environment variables before building:

```bash
source .env
docker build -t minitwitutmimage .
```

Start the local stack:

```bash
docker compose -f docker-compose-local.yml up -d
```

Stop the local stack:

```bash
docker compose -f docker-compose-local.yml down -v
```

The `-v` flag also removes volumes. Currently there is a single volume which persists database data even when postgres container is stopped.

Currently the project is hosted as a github release as well as a dockerhub image.

### Using docker image

A prerequisite is that you must have docker installed on your system in some way (either docker desktop or just the engine components).

- Download/pull the image

```bash
$ docker pull flakiator/minitwitimage
```

- Run the docker image  

There are a few flags here that are important -v will create a volume from your locally stored database if you have one if not it will just create an empty folder (the program intalizes an empty database if non exists so you should be fine to run it without having one locally). -p will bind the ports to a port locally on your machine. --rm and -d are described on the docker documentation website if you are interested

```bash
docker run -d -p 8080:8080 -v /tmp/minitwit.db:/tmp/minitwit.db --rm flakiator/minitwitimage
```

- You can now go to the website on localhost:8080

### Local binary install

- Go to the latest release of the program [Latest release](https://github.com/Slug-Boi-ITU-Repositories/Dense-Dino-Nuggets/releases/latest) (there at multiple versions that you can use depending on your system OS and CPU architecture)
- Once downloaded you can place the application in whatever location you prefer to run it from
- Run the application using  

```bash
$ ./minitwit
```

- You can now go to the website on localhost:8080

## Setup for developers

If you would like to build on this application you have to download `go` as well as have a `C` compiler to compile the flag tool used to flag tweets in the system. You will also need docker and vagrant if you would like to run the application in a container and provision it as a VM or on digital ocean

## Setup for Open Tofu

You need to have the tofu cli installed to run the commands below are the required tfvars.

```
do_token          = ""
postgres_user     = ""
postgres_password = ""
postgres_db       = ""
volume_mount      = ""
monitor_pub_key   = ""
postgres_host     = ""
monitor_ip        = ""
```

You also need to have these env vars set through an .env file.

```
AWS_ACCESS_KEY_ID=""
AWS_SECRET_ACCESS_KEY=""
TF_VAR_do_token=""
TF_VAR_postgres_user=""
TF_VAR_postgres_password=""
```

## Setup for vagrant

Required plugins for vagrant:

```
vagrant-digitalocean
vagrant-scp
vagrant-parallels
vagrant-reload
vagrant-vbguest
(for mac users)
vagrant_utm
```

Also if you want to mount your own db you need to put it in the directory:
`/tmp/minitwit/`

Set environment variables (remember to upload private ssh key to Digital ocean):
```
DIGITAL_OCEAN_TOKEN
SSH_KEY_NAME
DOCKER_USERNAME
```

And run mintwit application with either utm virtualbox or digital_ocean provider:

`vagrant up minitwit --provider=<provider>`

And run monitoring system with either utm, libvert, or digital_ocean provider:

`vagrant up monitoring --provider=<provider>`

## Running monitoring system
The monitoring system uses Prometheus and Grafana which is configured through the files in `./prometheus` and `./grafana`. The system can be started with the docker compose file in the repository.

### Running the monitoring for local testing and development
Build an image of the minitwit application using the Dockerfile in the repo with the tag `minitwit-monitoring`
```bash
$ docker build -t minitwit-monitoring .
```

Run the docker compose file with the local profile
```bash
$ docker compose --profile local up -d
```

Then the minitwit application is available on `localhost:8080`, Prometheus at `localhost:9090`, and Grafana at `localhost:3000`.

### Running the monitoring for production
This setup assumes that the minitwit system is already running and accessable. 
Edit the `./prometheus/prometheus_prod.yml` file so that the target matches the running minitwit.

The production setup assumes that the folders `./prometheus_data` and `./grafana_data` exists locally in the current directory. Make sure to create these first and set the permissions like this:
```bash
$ sudo chown -R 65534:65534 ./prometheus_data
$ sudo chown -R 472:472 ./grafana_data
```

Then the production setup can be started with
```bash
$ docker compose --profile prod up -d
```

## Running dagger workflows

Make sure the dagger engine is running

From the root of the project open dagger:
```bash
$ dagger
```

**The workflows that can be run:**

**Build**. *For building binaries for linux and mac*

```bash
$ build --src=.
```

**Test**. *For testing the whole system*
*This will test, lint and spellcheck the entire system*

```bash
$ dagger check
```

**Publish**. *Building and publishing docker image to DockerHub*
*Here DOCKER_PASSWORD refers the name of the env variable name that stores the password and not the actual password*
```bash
$ dagger call publish --src . --username "flakiator" --password "DOCKER_PASSWORD"
```
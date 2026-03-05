# Dense-Dino-Nuggets - Minitwit

This is a fork of a student project of a small twitter like application (this is lore). Currently this is a golang server application that can be used to boot up a small webpage where people can, create accounts, login and post chirps. There is a also logic for following and unfollowing other users so you can get your own custom timeline to look at.

## Installation

Currently the project is hosted as a github release as well as a dockerhub image.

### Using docker image

A prerequiste is that you must have docker installed on your system in some way (either docker desktop or just the engine components).

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

And run with either utm virtualbox or digital_ocean provider:

`vagrant up --provider=<provider>`
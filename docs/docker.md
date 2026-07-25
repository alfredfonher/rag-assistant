# Docker

Docker is an open-source platform that automates the deployment, scaling, and management of applications using containerization. It packages applications and their dependencies into lightweight, portable containers.

## What is Containerization?

Containerization is an operating-system-level virtualization method for running multiple isolated Linux systems on a control host using a single Linux kernel. Containers share the host OS kernel but have their own filesystem, CPU, memory, and process space.

## Docker Architecture

### Docker Engine
- **Docker Daemon (dockerd)**: Background service managing containers, images, networks, and volumes
- **REST API**: Interface for programs to communicate with the daemon
- **CLI (docker)**: Command-line interface for user interaction

### Docker Images
Images are read-only templates used to create containers. They are built in layers, with each layer representing a set of filesystem changes. Images are stored in registries like Docker Hub.

### Docker Containers
Containers are runnable instances of images. They are isolated from each other and the host system, with their own network interface, IP address, and process space.

### Dockerfile
A text file containing instructions to build a Docker image. Common instructions include:
- `FROM`: Base image
- `COPY`/`ADD`: Copy files into image
- `RUN`: Execute commands during build
- `CMD`/`ENTRYPOINT`: Default command when container starts
- `EXPOSE`: Document listening ports

### Docker Compose
A tool for defining and running multi-container applications using a YAML file. It manages the complete lifecycle of multi-service applications.

## Key Concepts

### Volumes
Persistent data storage that survives container restarts. Volumes can be named or anonymous and are managed by Docker.

### Networks
Docker provides multiple network drivers:
- **bridge**: Default network for single-host communication
- **host**: Remove network isolation between container and host
- **overlay**: Multi-host networking for Swarm services
- **none**: Disable all networking

### Images vs Containers
- Images are templates (class analogy)
- Containers are running instances (object analogy)
- Multiple containers can be created from a single image

## Docker Hub

The default public registry for Docker images, containing millions of images from official sources and community contributors. Users can push their own images and pull images from the registry.

## Benefits

- **Consistency**: Same environment from development to production
- **Isolation**: Applications don't interfere with each other
- **Portability**: Run anywhere Docker is installed
- **Efficiency**: Lightweight compared to virtual machines
- **Scalability**: Easy to scale horizontally
- **Version control**: Track changes to application environment

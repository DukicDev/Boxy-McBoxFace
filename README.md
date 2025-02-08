# Boxy-McBoxFace

[![Go Version](https://img.shields.io/github/go-mod/go-version/DukicDev/godoist)](https://golang.org/)  
[![License](https://img.shields.io/github/license/DukicDev/godoist)](LICENSE)

Boxy-McBoxFace is a hobby project written in Go that explores containerization, OCI images, and Linux isolation techniques. It’s a minimal container runtime that currently supports only Alpine Linux. 

## Overview

Boxy-McBoxFace demonstrates how to:
- Extract an OCI image tarball (generated via `docker save`) to create a container filesystem.
- Use Linux namespaces and `chroot` to isolate the container.
- Set up basic cgroup resource limits using [containerd/cgroups](https://github.com/containerd/cgroups).
- Run a simple container (currently only Alpine Linux is supported).

## Prerequisites

- **Go** (version 1.16 or later is recommended)
- **Docker** (to pull and save the Alpine image)
- A Linux system with support for namespaces, `chroot`, and cgroups

## Getting Started

### 1. Save the Alpine Image

Pull the Alpine image with Docker and save it locally:

```bash
docker pull alpine
docker save alpine -o ./images/alpine.tar

```

Make sure the tarball is saved in the `./images/` directory where you run Boxy-McBoxFace.

### 2. Build Boxy-McBoxFace

Clone the repository and build the project:

```bash
git clone https://github.com/DukicDev/boxy-mcboxface.git
cd boxy-mcboxface
go build -o boxy-mcboxface .
```

### 3. Run the Container

Run Boxy-McBoxFace using the following command:

```bash
sudo ./boxy-mcboxface run alpine
```

This will:
- Create a temporary container filesystem under `./boxy-mcboxface/alpine`
- Extract the Alpine OCI image into that directory
- Set up Linux namespaces, cgroups, and `chroot` into the new filesystem
- Execute the default command (typically `/bin/sh`) inside the container

### 4. Interact and Exit

Once the container is running, you should see a shell prompt. You can explore the Alpine environment. When you're done, type `exit` to leave the container. Boxy-McBoxFace is designed to clean up the extracted filesystem after the container stops.

### 5. Cleanup

If for any reason the cleanup doesn’t occur automatically, you can remove the container filesystem manually:

```bash
rm -rf ./boxy-mcboxface
```

## Project Structure

- **main.go:**  
  Contains the main entry point, command-line parsing, and logic for running the container (handling namespaces, cgroups, `chroot`, etc.).

- **imageHandler.go**  
  Responsible for extracting the OCI image tarball into the container filesystem.

## Future Directions

- **Support for More Images:**  
  Right now, only Alpine is supported. Future work might include support for additional images.

- **Enhanced Features:**  
  Potential improvements include better networking, storage options, more advanced resource management, and robust error handling.

## License

This project is licensed under the [MIT License](LICENSE).


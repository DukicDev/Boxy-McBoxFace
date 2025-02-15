# Boxy-McBoxFace

[![Go Version](https://img.shields.io/github/go-mod/go-version/DukicDev/godoist)](https://golang.org/)  
[![License](https://img.shields.io/github/license/DukicDev/godoist)](LICENSE)

Boxy-McBoxFace is a hobby project written in Go that explores containerization, OCI images, and Linux isolation techniques.

## Overview

Boxy-McBoxFace demonstrates how to:
- Pull the image from the docker registry.
- Use Linux namespaces and `chroot` to isolate the container.
- Set up basic cgroup resource limits using [containerd/cgroups](https://github.com/containerd/cgroups).
- Run a simple container.

## Prerequisites

- **Go** (version 1.16 or later is recommended)
- A Linux system with support for namespaces, `chroot`, and cgroups


## Getting Started

### 1. Build Boxy-McBoxFace

Clone the repository and build the project:

```bash
git clone https://github.com/DukicDev/boxy-mcboxface.git
cd boxy-mcboxface
go build -o boxy-mcboxface .
```

### 2. Run the Container

Run Boxy-McBoxFace using the following command:

```bash
sudo ./boxy-mcboxface run (imageName) (cmd)
```

This will:
- Create a temporary container filesystem under `/var/lib/boxy-mcboxface/containers/imageName`
- Pull and extract the OCI image into that directory
- Cache Layer files in `/var/lib/boxy-mcboxface/layers`
- Set up Linux namespaces, cgroups, and `chroot` into the new filesystem
- Execute the either default command (or cmd if given) inside the container


### 4. Cleanup

If for any reason the cleanup doesn’t occur automatically, you can remove the container filesystem manually:

```bash
(sudo) rm -rf /var/lib/boxy-mcboxface/containers/(imageName)
```

## Project Structure

- **main.go:**  
  Contains the main entry point, command-line parsing, and logic for running the container (handling namespaces, cgroups, `chroot`, etc.).

- **imageHandler.go**  
  Responsible for pulling and extracting the OCI image into the container filesystem.

## License

This project is licensed under the [MIT License](LICENSE).


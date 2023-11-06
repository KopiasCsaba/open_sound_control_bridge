<h1>Table of contents</h1>

<hr>


# Video Backend

<!-- TOC -->
* [Video Backend](#video-backend)
  * [Overview](#overview)
* [Development](#development)
  * [Requirements](#requirements)
  * [Building](#building)
  * [Starting the development environment](#starting-the-development-environment)
    * [make dev_start](#make-devstart)
  * [Understanding the codebase](#understanding-the-codebase)
* [Using the app](#using-the-app)
  * [Specifying secrets](#specifying-secrets)
* [Production](#production)
  * [Runtime requirements](#runtime-requirements)
* [./app --help](#app---help)
<!-- TOC -->

## Overview

@TODO add stuff here.

# Development

## Requirements

* [docker](https://docs.docker.com/engine/install/)
    * buildx (included in official docker builds,e.g. the docker-buildx-plugin package)
    * compose (included in official docker builds,e.g. the docker-compose-plugin package)
* git
* make
* jq
* curl
* gzip

WARNING:

* In case of ubuntu or other linux distros, don't use the distro supported version of docker.

## Building

* To build the application, use the "make build" command. The results will be in the "[build](./build)" folder.
* To debug the build process, start two shells:
    1. make build_interactive
    2. make build_shell

## Starting the development environment

### make dev_start

This is the most important command. This starts the application in a development container.
The app is killed and restarted upon any filechange, therefore iterating on code is very effective.

In case if it is needed, "make dev_shell" can be used to access the container in a shell.

## Understanding the codebase

Learn more about the basics of this codebase [docs/development/index.md](docs/development/index.md)

# Using the app

@TODO `./app --help`.

TL; DR: The application reads env variables, all of them are starting with a "APP_" prefix.
Also, environment variables can be stored in a plaintext file, one per line in the format: KEY=VALUE.

In this case the KEY must not start with APP_. Supply this file(s) to the app with the "--envfile" argument(s).
[Here](docker/devenv/devenv.env) is an example env file.

## Specifying secrets

There is an example secret file at [secrets.env.example](docker/devenv/secrets.env.example).

To use sections that require secrets, rename it to "secrets.env" and change the values to their respective value.
The dev environment automatically merges "secrets.env" onto the [devenv.env](docker/devenv/devenv.env)

# Production

To get an example for a production container configuration of the docker image, see "app-prod-runner" in
the [docker-compose.yml](docker/docker-compose.yml)!

## Runtime requirements


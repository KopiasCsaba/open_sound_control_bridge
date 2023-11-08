# syntax=docker/dockerfile:1.2

# Multi stage build file.
#
# buildbase (ubuntu)
# buildbase -> build
# buildbase -> development


# =========================================================================================================
FROM --platform=$BUILDPLATFORM ubuntu:22.04 AS buildbase

# https://vsupalov.com/buildkit-cache-mount-dockerfile/
RUN rm -f /etc/apt/apt.conf.d/docker-clean # Preventing the base-os to delete apt-cache.

# Just to confirm buildx is working. It would fail around here otherwise.
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN (echo "Running on buildplatform: $BUILDPLATFORM, targetplatform: $TARGETPLATFORM" && arch) > /log


ENV DEBIAN_FRONTEND="noninteractive"
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Install dependencies for compiling & building
# =============================================================================
RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt \
    apt-get update && \
    apt-get install -y --no-install-recommends git make build-essential curl wget nano tzdata ca-certificates



# INSTALL GO
# =============================================================================

RUN curl -sL https://go.dev/dl/go1.20.7.linux-amd64.tar.gz -o /tmp/go.tar.gz && tar -C /usr/local -xzf /tmp/go.tar.gz

# Add GO to PATH
RUN echo 'export PATH="${PATH}:/usr/local/go/bin"' >> /etc/profile


WORKDIR /mnt



# =========================================================================================================
FROM buildbase AS development

WORKDIR /tmp


# Install dependencies for the dev environment
RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt apt-get update && \
    apt-get install -y --no-install-recommends inotify-tools file mc

WORKDIR /mnt

# Add user with the same ID as the host (when specifies UID and GID).
# This helps solving file permission issues between host/container.

ARG UID=1000
ARG GID=1000
ARG USER=container

RUN groupadd -g ${GID} ${USER} \
    && useradd -u ${UID} -g ${GID} -d /home/${USER} -s /bin/bash -m ${USER}

USER ${USER}
RUN echo 'export PATH="${PATH}:${HOME}/go/bin"' >> ${HOME}/.bashrc

RUN bash -l -c "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
RUN bash -l -c "go install mvdan.cc/gofumpt@latest"
RUN bash -l -c "go install github.com/volatiletech/sqlboiler/v4@latest"
RUN bash -l -c "go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest"

ENTRYPOINT ["/bin/bash","-l","-c", "./docker/rundev.sh"]












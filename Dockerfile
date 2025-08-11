# dbn-go Dockerfile
# Copyright (c) 2024 Neomantra Corp

ARG DBNGO_BUILD_BASE="golang"
ARG DBNGO_BUILD_TAG="1.23-bullseye"

ARG DBNGO_RUNTIME_BASE="debian"
ARG DBNGO_RUNTIME_TAG="bullseye-slim"

##################################################################################################
# Builder
##################################################################################################

FROM ${DBNGO_BUILD_BASE}:${DBNGO_BUILD_TAG} AS build

ARG DBNGO_BUILD_BASE="golang"
ARG DBNGO_BUILD_TAG="1.23-bullseye"

# Extract TARGETARCH from BuildKit
ARG TARGETARCH

RUN DEBIAN_FRONTEND=noninteractive apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    git

ARG TASKFILE_VERSION="v3.44.1"
RUN curl -fSL "https://github.com/go-task/task/releases/download/${TASKFILE_VERSION}/task_linux_${TARGETARCH}.deb" -o /tmp/task_linux.deb \
    && dpkg -i /tmp/task_linux.deb \
    && rm /tmp/task_linux.deb

ARG GINKO_VERSION="v2.23.4"
RUN go install "github.com/onsi/ginkgo/v2/ginkgo@${GINKO_VERSION}"

ADD . /src
WORKDIR /src

# Regular build for smoke-test
RUN mkdir -p bin && task go-build

# Unit Tests
RUN task go-test-no-api

# Labels
LABEL DBNGO_BUILD_BASE="${DBNGO_BUILD_BASE}"
LABEL DBNGO_BUILD_TAG="${DBNGO_BUILD_TAG}"

LABEL DBNGO_TARGET_ARCH="${TARGETARCH}"

LABEL GINKO_VERSION="${GINKO_VERSION}"


##################################################################################################
# Runtime environment
###########################q#######################################################################

FROM ${DBNGO_RUNTIME_BASE}:${DBNGO_RUNTIME_TAG} AS runtime

ARG DBNGO_BUILD_BASE="golang"
ARG DBNGO_BUILD_TAG="1.23-bullseye"

ARG DBNGO_RUNTIME_BASE="debian"
ARG DBNGO_RUNTIME_TAG="bullseye-slim"

# Extract TARGETARCH from BuildKit
ARG TARGETARCH

# Install dependencies and ops tools
RUN DEBIAN_FRONTEND=noninteractive apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        coreutils \
        curl \
        zstd \
    && rm -rf /var/lib/apt/lists/*

# Copy binaries
COPY --from=build /src/bin/* /usr/local/bin/

# Labels
LABEL DBNGO_BUILD_BASE="${DBNGO_BUILD_BASE}"
LABEL DBNGO_BUILD_TAG="${DBNGO_BUILD_TAG}"
LABEL DBNGO_RUNTIME_BASE="${DBNGO_RUNTIME_BASE}"
LABEL DBNGO_RUNTIME_TAG="${DBNGO_RUNTIME_TAG}"

LABEL DBNGO_TARGET_ARCH="${TARGETARCH}"

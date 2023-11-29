#
# The docker image is built in two stages.  First stage builds the
# oasis-node and oasis-rosetta-gateway binaries.
# Second stage prepares the execution environment and copies the
# two binaries over.
#

#
# Build stage
#
FROM golang:1.21 AS build

# Install prerequisites.
RUN apt-get update && apt-get install -y bzip2 libseccomp-dev

ARG CORE_BRANCH=v23.0.7
ARG CORE_GITHUB=https://github.com/oasisprotocol/oasis-core

ARG GATEWAY_BRANCH=master
ARG GATEWAY_GITHUB=https://github.com/oasisprotocol/oasis-rosetta-gateway

ARG JEMALLOC_VERSION=5.2.1
ARG JEMALLOC_CHECKSUM=34330e5ce276099e2e8950d9335db5a875689a4c6a56751ef3b1d8c537f887f6

ARG GENESIS_MAINNET_URL=https://github.com/oasisprotocol/mainnet-artifacts/releases/download/2023-11-29/genesis.json
ARG GENESIS_MAINNET_FILE_CHECKSUM=b14e45e97da0216a16c25096fd216f591e4d526aa6abac110ac23cb327b64ba1

ARG GENESIS_TESTNET_URL=https://github.com/oasisprotocol/testnet-artifacts/releases/download/2023-10-12/genesis.json
ARG GENESIS_TESTNET_FILE_CHECKSUM=02ce385c050b2a5c7cf0e5e34f5e4930f7804bb21efba2d1d3aa8215123aab68

# Fetch and build jemalloc (used by BadgerDB).
RUN wget -O jemalloc.tar.bz2 \
    https://github.com/jemalloc/jemalloc/releases/download/${JEMALLOC_VERSION}/jemalloc-${JEMALLOC_VERSION}.tar.bz2 && \
    # Ensure checksum matches.
    echo "${JEMALLOC_CHECKSUM}  jemalloc.tar.bz2" | sha256sum -c && \
    tar -xf jemalloc.tar.bz2 && \
    cd jemalloc-${JEMALLOC_VERSION} && \
    # Ensure reproducible jemalloc build.
    # https://reproducible-builds.org/docs/build-path/
    EXTRA_CXXFLAGS=-ffile-prefix-map=$(pwd -L)=. EXTRA_CFLAGS=-ffile-prefix-map=$(pwd -L)=. \
    ./configure --with-jemalloc-prefix='je_' --with-malloc-conf='background_thread:true,metadata_thp:auto' && \
    make && \
    make install && \
    cd .. && rm jemalloc.tar.bz2 && rm -rf jemalloc-${JEMALLOC_VERSION}

# Fetch and build Oasis Core.
RUN git clone --single-branch --branch $CORE_BRANCH ${CORE_GITHUB} /usr/local/build-core && \
    cd /usr/local/build-core/go && \
    make oasis-node && \
    cp oasis-node/oasis-node /usr/bin/ && \
    make clean && go clean -cache -testcache -modcache && \
    cd / && rm -rf /usr/local/build-core

# Fetch and build Oasis Rosetta Gateway.
RUN git clone --single-branch --branch $GATEWAY_BRANCH ${GATEWAY_GITHUB} /usr/local/build-gateway && \
    cd /usr/local/build-gateway && \
    make && \
    cp oasis-rosetta-gateway /usr/bin/ && \
    mkdir -p /usr/local/misc-gateway && \
    cp -r \
        docker/service \
        docker/config.yml \
        docker/config-testnet.yml \
        /usr/local/misc-gateway/ && \
    make nuke && \
    cd / && rm -rf /usr/local/build-gateway

# Fetch Oasis genesis documents.
RUN wget -O /usr/local/misc-gateway/genesis.json \
    "$GENESIS_MAINNET_URL" && \
    # Ensure checksum matches.
    echo "$GENESIS_MAINNET_FILE_CHECKSUM /usr/local/misc-gateway/genesis.json" | sha256sum -c && \
    wget -O /usr/local/misc-gateway/genesis-testnet.json \
    "$GENESIS_TESTNET_URL" && \
    # Ensure checksum matches.
    echo "$GENESIS_TESTNET_FILE_CHECKSUM /usr/local/misc-gateway/genesis-testnet.json" | sha256sum -c


#
# Execution stage
#
FROM ubuntu:jammy
COPY --from=build /usr/bin/oasis-node /usr/bin/
COPY --from=build /usr/bin/oasis-rosetta-gateway /usr/bin/
RUN mkdir -m700 /data /data/etc /data/data && \
    mkdir -p /etc/service && \
    apt-get update && apt-get install -y runit
COPY --from=build /usr/local/misc-gateway/genesis.json /data/etc/
COPY --from=build /usr/local/misc-gateway/genesis-testnet.json /data/etc/
COPY --from=build /usr/local/misc-gateway/config.yml /data/etc/
COPY --from=build /usr/local/misc-gateway/config-testnet.yml /data/etc/
COPY --from=build /usr/local/misc-gateway/service /etc/service/

# Prepare for non-root oasis-node.
RUN chown -R www-data:www-data /data

VOLUME /data
WORKDIR /data
ENV OASIS_NODE_GRPC_ADDR="unix:/data/internal.sock"
ENV OASIS_NODE_CONFIG="/data/etc/config.yml"

# Start the node and the rosetta-gateway.
CMD runsvdir /etc/service

# Expose gateway and node ports.
EXPOSE 8080/tcp
EXPOSE 26656/tcp

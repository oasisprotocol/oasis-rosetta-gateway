#
# The docker image is built in two stages.  First stage builds the
# oasis-node and oasis-core-rosetta-gateway binaries.
# Second stage prepares the execution environment and copies the
# two binaries over.
#

#
# Build stage
#
FROM golang:1.16-alpine3.13 AS build

# Install prerequisites.
RUN rm -f /var/cache/apk/*; apk --no-cache add bash git make gcc g++ libressl-dev libseccomp-dev linux-headers

ARG CORE_BRANCH=v21.2.6
ARG CORE_GITHUB=https://github.com/oasisprotocol/oasis-core

ARG GATEWAY_BRANCH=master
ARG GATEWAY_GITHUB=https://github.com/oasisprotocol/oasis-core-rosetta-gateway

ARG JEMALLOC_VERSION=5.2.1
ARG JEMALLOC_CHECKSUM=34330e5ce276099e2e8950d9335db5a875689a4c6a56751ef3b1d8c537f887f6

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

# Fetch and build Oasis Core Rosetta Gateway.
RUN git clone --single-branch --branch $GATEWAY_BRANCH ${GATEWAY_GITHUB} /usr/local/build-gateway && \
    cd /usr/local/build-gateway && \
    make && \
    cp oasis-core-rosetta-gateway /usr/bin/ && \
    make nuke && \
    cd / && rm -rf /usr/local/build-gateway


#
# Execution stage
#
FROM alpine:3.13
RUN rm -f /var/cache/apk/*; apk --no-cache add bash libressl libseccomp
COPY --from=build /usr/bin/oasis-node /usr/bin/
COPY --from=build /usr/bin/oasis-core-rosetta-gateway /usr/bin/

VOLUME /data
WORKDIR /data
ENV OASIS_NODE_GRPC_ADDR="unix:/data/internal.sock"

# Start the node and the rosetta-gateway.
CMD oasis-node --config /data/etc/config.yml & oasis-core-rosetta-gateway

# Expose gateway and node ports.
EXPOSE 8080/tcp
EXPOSE 26656/tcp

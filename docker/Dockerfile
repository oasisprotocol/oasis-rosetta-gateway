#
# The docker image is built in two stages.  First stage builds the
# oasis-node and oasis-core-rosetta-gateway binaries.
# Second stage prepares the execution environment and copies the
# two binaries over.
#

#
# Build stage
#
FROM golang:1.14-alpine AS build

# Install prerequisites.
RUN rm -f /var/cache/apk/*; apk --no-cache add bash git make gcc g++ libressl-dev libseccomp-dev linux-headers

ARG CORE_BRANCH="v20.7"
ARG GATEWAY_BRANCH="master"

# Fetch and build oasis-core.
RUN git clone --single-branch --branch $CORE_BRANCH https://github.com/oasisprotocol/oasis-core /usr/local/build-core && \
	cd /usr/local/build-core/go && \
	make oasis-node && \
	cp oasis-node/oasis-node /usr/bin/ && \
	make clean && go clean -cache -testcache -modcache && \
	cd / && rm -rf /usr/local/build-core

# Fetch and build oasis-core-rosetta-gateway.
RUN git clone --single-branch --branch $GATEWAY_BRANCH https://github.com/oasisprotocol/oasis-core-rosetta-gateway /usr/local/build-gateway && \
	 cd /usr/local/build-gateway && \
	 make && \
	 cp oasis-core-rosetta-gateway /usr/bin/ && \
	 make nuke && \
	 cd / && rm -rf /usr/local/build-gateway


#
# Execution stage
#
FROM alpine:latest
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

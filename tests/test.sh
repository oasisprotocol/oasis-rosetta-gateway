#!/usr/bin/env bash
set -o nounset -o pipefail -o errexit
trap "exit 1" INT

# Get the root directory of the tests dir inside the repository.
ROOT="$(cd $(dirname $0); pwd -P)"
cd "${ROOT}"

# ANSI escape codes to brighten up the output.
GRN=$'\e[32;1m'
OFF=$'\e[0m'

# Paths to various binaries and config files that we need.
OASIS_ROSETTA_GW="${ROOT}/../oasis-core-rosetta-gateway"
OASIS_NET_RUNNER="${ROOT}/oasis-net-runner"
OASIS_NODE="${ROOT}/oasis-node"
FIXTURE_FILE="${ROOT}/test-fixture-config.json"

# Destination address for test transfers.
DST="oasis1qpkant39yhx59sagnzpc8v0sg8aerwa3jyqde3ge"

# Kill all dangling processes on exit.
cleanup() {
	printf "${OFF}"
	pkill -P $$ || true
	wait || true
}
trap "cleanup" EXIT

# The base directory for all the node and test env cruft.
TEST_BASE_DIR=$(mktemp -d -t oasis-rosetta-XXXXXXXXXX)

# The oasis-node binary must be in the path for the oasis-net-runner to find it.
export PATH="${PATH}:${ROOT}"

printf "${GRN}### Starting the test network...${OFF}\n"
${OASIS_NET_RUNNER} \
	--fixture.file ${FIXTURE_FILE} \
	--basedir.no_temp_dir \
	--basedir ${TEST_BASE_DIR} &

export OASIS_NODE_GRPC_ADDR="unix:${TEST_BASE_DIR}/net-runner/network/client-0/internal.sock"

# How many nodes to wait for each epoch.
NUM_NODES=1

# Current nonce for transactions (incremented after every submit_tx).
NONCE=0

# Helper function for advancing the current epoch to the given parameter.
advance_epoch() {
	local epoch=$1
	printf "${GRN}### Advancing epoch ($epoch)...${OFF}\n"
	${OASIS_NODE} debug control set-epoch \
		--address ${OASIS_NODE_GRPC_ADDR} \
		--epoch $epoch
}

# Helper function that waits for all nodes to register.
wait_for_nodes() {
	printf "${GRN}### Waiting for all nodes to register...${OFF}\n"
	${OASIS_NODE} debug control wait-nodes \
		--address ${OASIS_NODE_GRPC_ADDR} \
		--nodes ${NUM_NODES} \
		--wait
}

# Helper function that submits the given transaction JSON file.
submit_tx() {
	local tx=$1
	# Submit transaction.
	${OASIS_NODE} consensus submit_tx \
		--transaction.file "$tx" \
		--address ${OASIS_NODE_GRPC_ADDR} \
		--debug.allow_test_keys
	# Increase nonce.
	NONCE=$((NONCE+1))
}

# Helper function that generates a transfer transaction.
gen_transfer() {
	local tx=$1
	local amount=$2
	local dst=$3
	${OASIS_NODE} stake account gen_transfer \
		--stake.amount $amount \
		--stake.transfer.destination "$dst" \
		--transaction.file "$tx" \
		--transaction.nonce ${NONCE} \
		--transaction.fee.amount 0 \
		--transaction.fee.gas 10000 \
		--debug.dont_blame_oasis \
		--debug.test_entity \
		--debug.allow_test_keys \
		--genesis.file "${TEST_BASE_DIR}/net-runner/network/genesis.json"
}

# Helper function that generates a burn transaction.
gen_burn() {
	local tx=$1
	local amount=$2
	${OASIS_NODE} stake account gen_burn \
		--stake.amount $amount \
		--transaction.file "$tx" \
		--transaction.nonce ${NONCE} \
		--transaction.fee.amount 1 \
		--transaction.fee.gas 10000 \
		--debug.dont_blame_oasis \
		--debug.test_entity \
		--debug.allow_test_keys \
		--genesis.file "${TEST_BASE_DIR}/net-runner/network/genesis.json"
}

printf "${GRN}### Waiting for the validator to register...${OFF}\n"
${OASIS_NODE} debug control wait-nodes \
	--address ${OASIS_NODE_GRPC_ADDR} \
	--nodes 1 \
	--wait

advance_epoch 1
wait_for_nodes

printf "${GRN}### Burning tokens...${OFF}\n"
gen_burn "${TEST_BASE_DIR}/burn.json" 42
submit_tx "${TEST_BASE_DIR}/burn.json"

advance_epoch 2
wait_for_nodes

printf "${GRN}### Transferring tokens (1)...${OFF}\n"
gen_transfer "${TEST_BASE_DIR}/tx1.json" 1000 "${DST}"
submit_tx "${TEST_BASE_DIR}/tx1.json"

advance_epoch 3
wait_for_nodes

printf "${GRN}### Transferring tokens (2)...${OFF}\n"
gen_transfer "${TEST_BASE_DIR}/tx2.json" 123 "${DST}"
submit_tx "${TEST_BASE_DIR}/tx2.json"

printf "${GRN}### Transferring tokens (3)...${OFF}\n"
gen_transfer "${TEST_BASE_DIR}/tx3.json" 456 "${DST}"
submit_tx "${TEST_BASE_DIR}/tx3.json"

advance_epoch 4
wait_for_nodes

advance_epoch 5
wait_for_nodes

advance_epoch 6
wait_for_nodes

printf "${GRN}### Starting the Rosetta gateway...${OFF}\n"
${OASIS_ROSETTA_GW} &

sleep 3

printf "${GRN}### Validating Rosetta gateway implementation...${OFF}\n"
go run ./check-prep
./rosetta-cli --configuration-file rosetta-cli-config.json check:data --end 42
rm -rf "${ROOT}/validator-data" /tmp/rosetta-cli*

printf "${GRN}### Testing construction signing workflow...${OFF}\n"
go run ./construction-signing

printf "${GRN}### Testing construction transaction types...${OFF}\n"
go run ./construction-txtypes

# Clean up after a successful run.
rm -rf "${TEST_BASE_DIR}"

printf "${GRN}### Tests finished.${OFF}\n"

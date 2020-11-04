#!/usr/bin/env bash
set -o nounset -o pipefail -o errexit
trap "exit 1" INT

# Get the root directory of the tests dir inside the repository.
ROOT="$(cd $(dirname $0); pwd -P)"
cd "${ROOT}"

# ANSI escape codes to brighten up the output.
RED=$'\e[31;1m'
GRN=$'\e[32;1m'
OFF=$'\e[0m'

# Paths to various binaries and config files that we need.
OASIS_ROSETTA_GW="${ROOT}/../oasis-core-rosetta-gateway"
OASIS_NET_RUNNER="${ROOT}/oasis-net-runner"
OASIS_NODE="${ROOT}/oasis-node"

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

# Helper function for running the test network.
start_network() {
	local height=$1
	${OASIS_NET_RUNNER} \
	    --fixture.default.node.binary ${OASIS_NODE} \
		--fixture.default.initial_height=${height} \
		--fixture.default.setup_runtimes=false \
		--fixture.default.num_entities=1 \
		--fixture.default.epochtime_mock=true \
		--basedir.no_temp_dir \
		--basedir ${TEST_BASE_DIR} &

}

printf "${GRN}### Starting the test network...${OFF}\n"
start_network 1

export OASIS_NODE_GRPC_ADDR="unix:${TEST_BASE_DIR}/net-runner/network/validator-0/internal.sock"

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
		--assume_yes \
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
		--assume_yes \
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
${OASIS_GO} run ./check-prep
./rosetta-cli --configuration-file rosetta-cli-config.json check:data --end 42
{
  # We'll cause a sigpipe on this process, so ignore the exit status.
  # The downstream awk will exit with nonzero status if this test actually fails without confirming any transactions.
  ./rosetta-cli --configuration-file rosetta-cli-config.json check:construction || true
} | stdbuf -oL awk '{ print $0 }; $1 == "[STATS]" && $4 >= 42 { confirmed = 1; exit }; END { exit !confirmed }'
rm -rf "${ROOT}/validator-data" /tmp/rosetta-cli*

printf "${GRN}### Testing construction signing workflow...${OFF}\n"
${OASIS_GO} run ./construction-signing

printf "${GRN}### Testing construction transaction types...${OFF}\n"
${OASIS_GO} run ./construction-txtypes

# Now test if the initial block height change works on a new network.
printf "${GRN}### Terminating existing test network...${OFF}\n"
cleanup
rm -rf "${TEST_BASE_DIR}"
mkdir -p "${TEST_BASE_DIR}"

printf "${GRN}### Starting new test network at non-1 height...${OFF}\n"
start_network 123
NONCE=0

sleep 2

advance_epoch 1
wait_for_nodes

printf "${GRN}### Starting the Rosetta gateway (again)...${OFF}\n"
${OASIS_ROSETTA_GW} &

sleep 3

printf "${GRN}### Transferring tokens (1+)...${OFF}\n"
gen_transfer "${TEST_BASE_DIR}/tx1.json" 1000 "${DST}"
submit_tx "${TEST_BASE_DIR}/tx1.json"

advance_epoch 2
wait_for_nodes

advance_epoch 3
wait_for_nodes

advance_epoch 4
wait_for_nodes


printf "${GRN}### Validating Rosetta gateway implementation (again)...${OFF}\n"
${OASIS_GO} run ./check-prep
./rosetta-cli --configuration-file rosetta-cli-config.json check:data --end 135

# Clean up after a successful run.
printf "${GRN}### Terminating existing test network...${OFF}\n"
cleanup
rm -rf "${TEST_BASE_DIR}" /tmp/rosetta-cli*
mkdir -p "${TEST_BASE_DIR}"

# Test offline mode.
unset OASIS_NODE_GRPC_ADDR
export OASIS_ROSETTA_GATEWAY_OFFLINE_MODE="1"
export OASIS_ROSETTA_GATEWAY_OFFLINE_MODE_CHAIN_ID="test"

printf "${GRN}### Starting the Rosetta gateway in offline mode...${OFF}\n"
${OASIS_ROSETTA_GW} &

sleep 1

printf "${GRN}### Testing /construction/derive in offline mode...${OFF}\n"
OUTPUT=$(curl -s -H 'Content-Type: application/json' -X POST \
	-d '{"network_identifier":{"blockchain":"Oasis","network":"test"},"public_key":{"hex_bytes":"1234567890000000000000000000000000000000000000000000000000000000","curve_type":"edwards25519"}}' \
	http://localhost:8080/construction/derive \
	| \
	fgrep 'oasis1qp7cahykn900m3pxsnq7xw0zgvcuul0wtcpyrlp6')
if [[ "${OUTPUT}" != '{"address":"oasis1qp7cahykn900m3pxsnq7xw0zgvcuul0wtcpyrlp6"}' ]]; then
	printf "${RED}FAILURE${OFF}\n"
	exit 1
else
	printf "${GRN}SUCCESS${OFF}\n"
fi

rm -rf "${TEST_BASE_DIR}" /tmp/rosetta-cli*
printf "${GRN}### Tests finished.${OFF}\n"

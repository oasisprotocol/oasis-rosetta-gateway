package services

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"

	oc "github.com/oasisprotocol/oasis-core-rosetta-gateway/oasis-client"
	"github.com/oasisprotocol/oasis-core/go/common/logging"
)

var loggerNet = logging.GetLogger("services/network")

type networkAPIService struct {
	oasisClient oc.OasisClient
}

// NewNetworkAPIService creates a new instance of a NetworkAPIService.
func NewNetworkAPIService(oasisClient oc.OasisClient) server.NetworkAPIServicer {
	return &networkAPIService{
		oasisClient: oasisClient,
	}
}

// NetworkList implements the /network/list endpoint.
func (s *networkAPIService) NetworkList(
	ctx context.Context,
	request *types.MetadataRequest,
) (*types.NetworkListResponse, *types.Error) {
	chainID, err := GetChainID(ctx, s.oasisClient)
	if err != nil {
		loggerNet.Error("NetworkList: unable to get chain ID")
		return nil, err
	}

	resp := &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			&types.NetworkIdentifier{
				Blockchain: OasisBlockchainName,
				Network:    chainID,
			},
		},
	}

	jr, _ := json.Marshal(resp)
	loggerNet.Debug("NetworkList OK", "response", jr)

	return resp, nil
}

// NetworkStatus implements the /network/status endpoint.
func (s *networkAPIService) NetworkStatus(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.NetworkStatusResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerNet.Error("NetworkStatus: network validation failed", "err", terr.Message)
		return nil, terr
	}

	status, err := s.oasisClient.GetStatus(ctx)
	if err != nil {
		loggerNet.Error("NetworkStatus: unable to get node status", "err", err)
		return nil, ErrUnableToGetNodeStatus
	}

	peers := []*types.Peer{}
	for _, p := range status.Consensus.NodePeers {
		peers = append(peers, &types.Peer{
			PeerID: p,
		})
	}

	resp := &types.NetworkStatusResponse{
		CurrentBlockIdentifier: &types.BlockIdentifier{
			Index: status.Consensus.LatestHeight,
			Hash:  hex.EncodeToString(status.Consensus.LatestHash),
		},
		CurrentBlockTimestamp: status.Consensus.LatestTime.UnixNano() / 1000000, // ms
		GenesisBlockIdentifier: &types.BlockIdentifier{
			Index: status.Consensus.GenesisHeight,
			Hash:  hex.EncodeToString(status.Consensus.GenesisHash),
		},
		Peers: peers,
	}

	jr, _ := json.Marshal(resp)
	loggerNet.Debug("NetworkStatus OK", "response", jr)

	return resp, nil
}

// NetworkOptions implements the /network/options endpoint.
func (s *networkAPIService) NetworkOptions(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.NetworkOptionsResponse, *types.Error) {
	terr := ValidateNetworkIdentifier(ctx, s.oasisClient, request.NetworkIdentifier)
	if terr != nil {
		loggerNet.Error("NetworkStatus: network validation failed", "err", terr.Message)
		return nil, terr
	}

	status, err := s.oasisClient.GetStatus(ctx)
	if err != nil {
		loggerNet.Error("NetworkStatus: unable to get node status", "err", err)
		return nil, ErrUnableToGetNodeStatus
	}

	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion: "1.3.5",
			NodeVersion:    status.SoftwareVersion,
		},
		Allow: &types.Allow{
			OperationStatuses: []*types.OperationStatus{
				{
					Status:     OpStatusOK,
					Successful: true,
				},
			},
			OperationTypes: []string{
				OpTransfer,
				OpBurn,
			},
			Errors: ErrorList,
		},
	}, nil
}

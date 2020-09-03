package services

import (
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction/results"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
)

const (
	// OpTransfer is the Transfer operation.
	OpTransfer = "Transfer"
	// OpBurn is the Burn operation.
	OpBurn = "Burn"
	// OpReclaimEscrow is the Burn operation.
	OpReclaimEscrow = "ReclaimEscrow"
)

// SupportedOperationTypes is a list of the supported operations.
var SupportedOperationTypes = []string{
	OpTransfer,
	OpBurn,
	OpReclaimEscrow,
}

const (
	// OpStatusOK is the operation status for successful operations.
	OpStatusOK = "OK"
	// OpStatusFailed is the operation status for failed operations.
	OpStatusFailed = "Failed"
)

type transactionsDecoder struct {
	txs   []*types.Transaction
	index map[hash.Hash]*types.Transaction
}

func (d *transactionsDecoder) DecodeTx(rawTx []byte, result *results.Result) error {
	var sigTx transaction.SignedTransaction
	if err := cbor.Unmarshal(rawTx, &sigTx); err != nil {
		return fmt.Errorf("malformed transaction: %w", err)
	}
	var tx transaction.Transaction
	if err := sigTx.Open(&tx); err != nil {
		return fmt.Errorf("bad transaction signature: %w", err)
	}

	txHash := sigTx.Hash()
	rosettaTx := d.getOrCreateTx(txHash)

	// Decode events emitted by the transaction.
	d.decodeEvents(rosettaTx, result.Events)
	// In case this transaction failed, there were no events emitted for the failing parts.
	//
	// Case 1:
	// * Fee operations were already emitted when processing events as the fee was successfully
	//   deducted before processing the transaction.
	// * The transaction failed so no events were emitted and so we need to generate Failed
	//   operations.
	//
	// Case 2:
	// * Fee operations were not emitted when processing events, for example, because there was not
	//   enough balance in the account. Failed fee operations need to be emitted here.
	// * The transaction also failed in this case (since it got aborted when processing the fee),
	//   and so we need to generate Failed operations.
	if !result.IsSuccess() {
		txSignerAddress := StringFromAddress(staking.NewAddress(sigTx.Signature.PublicKey))
		o2t := newOperationToTransactionMapper(rosettaTx.Operations)
		t2o := newTransactionToOperationMapper(&tx, txSignerAddress, OpStatusFailed, rosettaTx.Operations)

		// If no fee operations were emitted, emit some now.
		if !o2t.HasFee() {
			t2o.EmitFeeOps()
		}
		if err := t2o.EmitTxOps(); err != nil {
			return fmt.Errorf("bad transaction: %w", err)
		}

		rosettaTx.Operations = t2o.Operations()
	}
	return nil
}

func (d *transactionsDecoder) DecodeBlock(blkHash hash.Hash, events []*staking.Event) error {
	for _, ev := range events {
		// We put all block-level events under an empty "transaction". All other events are skipped
		// as they have already been processed during DecodeTx.
		if !ev.TxHash.IsEmpty() {
			continue
		}

		rosettaTx := d.getOrCreateTx(blkHash)
		d.decodeStakingEvent(rosettaTx, ev)
	}
	return nil
}

func (d *transactionsDecoder) Transactions() []*types.Transaction {
	return d.txs
}

func (d *transactionsDecoder) getOrCreateTx(txHash hash.Hash) *types.Transaction {
	if tx, exists := d.index[txHash]; exists {
		return tx
	}

	tx := &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txHash.String(),
		},
		Operations: []*types.Operation{},
	}

	d.txs = append(d.txs, tx)
	d.index[txHash] = tx
	return tx
}

func (d *transactionsDecoder) decodeEvents(tx *types.Transaction, events []*results.Event) {
	for _, ev := range events {
		// We are only interested in staking events.
		if ev.Staking == nil {
			continue
		}

		d.decodeStakingEvent(tx, ev.Staking)
	}
}

func appendOp(ops []*types.Operation, kind string, acct string, subacct *types.SubAccountIdentifier, amt string) []*types.Operation {
	opIndex := int64(len(ops))
	op := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: opIndex,
		},
		Type:   kind,
		Status: OpStatusOK,
		Account: &types.AccountIdentifier{
			Address:    acct,
			SubAccount: subacct,
		},
		Amount: &types.Amount{
			Value:    amt,
			Currency: OasisCurrency,
		},
	}

	// Add related operation if it exists.
	if opIndex >= 1 {
		op.RelatedOperations = []*types.OperationIdentifier{
			&types.OperationIdentifier{
				Index: opIndex - 1,
			},
		}
	}

	return append(ops, op)
}

func (d *transactionsDecoder) decodeStakingEvent(tx *types.Transaction, ev *staking.Event) {
	switch {
	case ev.Transfer != nil:
		tx.Operations = appendOp(tx.Operations, OpTransfer, StringFromAddress(ev.Transfer.From), nil, "-"+ev.Transfer.Amount.String())
		tx.Operations = appendOp(tx.Operations, OpTransfer, StringFromAddress(ev.Transfer.To), nil, ev.Transfer.Amount.String())
	case ev.Burn != nil:
		tx.Operations = appendOp(tx.Operations, OpBurn, StringFromAddress(ev.Burn.Owner), nil, "-"+ev.Burn.Amount.String())
	case ev.Escrow != nil:
		ee := ev.Escrow
		switch {
		case ee.Add != nil:
			// Owner's general account -> escrow account.
			tx.Operations = appendOp(tx.Operations, OpTransfer, StringFromAddress(ee.Add.Owner), nil, "-"+ee.Add.Amount.String())
			tx.Operations = appendOp(tx.Operations, OpTransfer, StringFromAddress(ee.Add.Escrow), &types.SubAccountIdentifier{Address: SubAccountEscrow}, ee.Add.Amount.String())
		case ee.Take != nil:
			tx.Operations = appendOp(tx.Operations, OpTransfer, StringFromAddress(ee.Take.Owner), &types.SubAccountIdentifier{Address: SubAccountEscrow}, "-"+ee.Take.Amount.String())
			tx.Operations = appendOp(tx.Operations, OpTransfer, StringFromAddress(staking.CommonPoolAddress), nil, ee.Take.Amount.String())
		case ee.Reclaim != nil:
			// Escrow account -> owner's general account.
			tx.Operations = appendOp(tx.Operations, OpTransfer, StringFromAddress(ee.Reclaim.Escrow), &types.SubAccountIdentifier{Address: SubAccountEscrow}, "-"+ee.Reclaim.Amount.String())
			tx.Operations = appendOp(tx.Operations, OpTransfer, StringFromAddress(ee.Reclaim.Owner), nil, ee.Reclaim.Amount.String())
		}
	}
}

func newTransactionsDecoder() *transactionsDecoder {
	return &transactionsDecoder{
		txs:   []*types.Transaction{},
		index: make(map[hash.Hash]*types.Transaction),
	}
}

type operationToTransactionMapper struct {
	ops []*types.Operation
}

// HasFee verifies whether the given operation list contains fee payment operations.
func (m *operationToTransactionMapper) HasFee() bool {
	return len(m.ops) >= 2 &&
		m.ops[0].Type == OpTransfer &&
		m.ops[1].Type == OpTransfer &&
		m.ops[1].Account.Address == StringFromAddress(staking.FeeAccumulatorAddress)
}

func newOperationToTransactionMapper(ops []*types.Operation) *operationToTransactionMapper {
	return &operationToTransactionMapper{ops}
}

type transactionToOperationMapper struct {
	tx              *transaction.Transaction
	txSignerAddress string
	status          string

	ops []*types.Operation
}

// Operations returns the mapped operations.
func (m *transactionToOperationMapper) Operations() []*types.Operation {
	return m.ops
}

// EmitFeeOps emits the required fee operations if needed.
func (m *transactionToOperationMapper) EmitFeeOps() {
	if m.tx.Fee == nil || m.tx.Fee.Amount.IsZero() {
		return
	}

	opIndex := int64(len(m.ops))
	feeAmountStr := m.tx.Fee.Amount.String()
	m.ops = append(m.ops,
		&types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: opIndex,
			},
			Type:   OpTransfer,
			Status: m.status,
			Account: &types.AccountIdentifier{
				Address: m.txSignerAddress,
			},
			Amount: &types.Amount{
				Value:    "-" + feeAmountStr,
				Currency: OasisCurrency,
			},
			Metadata: map[string]interface{}{
				FeeGasKey: m.tx.Fee.Gas,
			},
		},
		&types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: opIndex + 1,
			},
			Type:   OpTransfer,
			Status: m.status,
			Account: &types.AccountIdentifier{
				Address: StringFromAddress(staking.FeeAccumulatorAddress),
			},
			Amount: &types.Amount{
				Value:    feeAmountStr,
				Currency: OasisCurrency,
			},
			RelatedOperations: []*types.OperationIdentifier{
				&types.OperationIdentifier{
					Index: opIndex,
				},
			},
		},
	)
}

// EmitTxOps emits the required transaction-specific operations.
func (m *transactionToOperationMapper) EmitTxOps() error {
	opIndex := int64(len(m.ops))

	switch m.tx.Method {
	case staking.MethodTransfer:
		var body staking.Transfer
		if err := cbor.Unmarshal(m.tx.Body, &body); err != nil {
			return fmt.Errorf("malformed body: %w", err)
		}

		m.ops = append(m.ops,
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: opIndex,
				},
				Type:   OpTransfer,
				Status: m.status,
				Account: &types.AccountIdentifier{
					Address: m.txSignerAddress,
				},
				Amount: &types.Amount{
					Value:    "-" + body.Amount.String(),
					Currency: OasisCurrency,
				},
			},
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: opIndex + 1,
				},
				Type:   OpTransfer,
				Status: m.status,
				Account: &types.AccountIdentifier{
					Address: StringFromAddress(body.To),
				},
				Amount: &types.Amount{
					Value:    body.Amount.String(),
					Currency: OasisCurrency,
				},
				RelatedOperations: []*types.OperationIdentifier{
					&types.OperationIdentifier{
						Index: opIndex,
					},
				},
			},
		)
	case staking.MethodBurn:
		var body staking.Burn
		if err := cbor.Unmarshal(m.tx.Body, &body); err != nil {
			return fmt.Errorf("malformed body: %w", err)
		}

		m.ops = append(m.ops,
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: opIndex,
				},
				Type:   OpBurn,
				Status: m.status,
				Account: &types.AccountIdentifier{
					Address: m.txSignerAddress,
				},
				Amount: &types.Amount{
					Value:    "-" + body.Amount.String(),
					Currency: OasisCurrency,
				},
			},
		)
	case staking.MethodAddEscrow:
		var body staking.Escrow
		if err := cbor.Unmarshal(m.tx.Body, &body); err != nil {
			return fmt.Errorf("malformed body: %w", err)
		}

		m.ops = append(m.ops,
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: opIndex,
				},
				Type:   OpTransfer,
				Status: m.status,
				Account: &types.AccountIdentifier{
					Address: m.txSignerAddress,
				},
				Amount: &types.Amount{
					Value:    "-" + body.Amount.String(),
					Currency: OasisCurrency,
				},
			},
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: opIndex + 1,
				},
				Type:   OpTransfer,
				Status: m.status,
				Account: &types.AccountIdentifier{
					Address: StringFromAddress(body.Account),
					SubAccount: &types.SubAccountIdentifier{
						Address: SubAccountEscrow,
					},
				},
				Amount: &types.Amount{
					Value:    body.Amount.String(),
					Currency: OasisCurrency,
				},
				RelatedOperations: []*types.OperationIdentifier{
					&types.OperationIdentifier{
						Index: opIndex,
					},
				},
			},
		)
	case staking.MethodReclaimEscrow:
		var body staking.ReclaimEscrow
		if err := cbor.Unmarshal(m.tx.Body, &body); err != nil {
			return fmt.Errorf("malformed body: %w", err)
		}

		m.ops = append(m.ops,
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: opIndex,
				},
				Type:   OpReclaimEscrow,
				Status: m.status,
				Account: &types.AccountIdentifier{
					Address: m.txSignerAddress,
				},
			},
			&types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: opIndex + 1,
				},
				Type:   OpReclaimEscrow,
				Status: m.status,
				Account: &types.AccountIdentifier{
					Address: StringFromAddress(body.Account),
					SubAccount: &types.SubAccountIdentifier{
						Address: SubAccountEscrow,
					},
				},
				Metadata: map[string]interface{}{
					ReclaimEscrowSharesKey: body.Shares.String(),
				},
				RelatedOperations: []*types.OperationIdentifier{
					&types.OperationIdentifier{
						Index: opIndex,
					},
				},
			},
		)
	default:
		// Other transactions do not affect balances so they do not emit any operations.
	}
	return nil
}

func newTransactionToOperationMapper(
	tx *transaction.Transaction,
	txSignerAddress string,
	status string,
	ops []*types.Operation,
) *transactionToOperationMapper {
	return &transactionToOperationMapper{
		tx:              tx,
		txSignerAddress: txSignerAddress,
		status:          status,
		ops:             ops,
	}
}

package services

import (
	"fmt"
	"math/big"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/common/quantity"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction/results"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
)

// FeeGasKey is the name of the key in the Metadata map inside a fee
// operation that specifies the gas value in the transaction fee.
// This is optional, and we use DefaultGas if it's absent.
const FeeGasKey = "fee_gas"

// ReclaimEscrowSharesKey is the name of the key in the Metadata map inside a
// reclaim escrow operation that specifies the number of shares to reclaim.
const ReclaimEscrowSharesKey = "reclaim_escrow_shares"

// DefaultGas is the default gas limit used in creating a transaction.
const DefaultGas transaction.Gas = 10000

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

// GetFee returns the fee (if any) extracted from the fee payment operations.
//
// If fee payment operations are present, this method also returns the signer address.
func (m *operationToTransactionMapper) GetFee() (string, *transaction.Fee, error) {
	fee := transaction.Fee{
		Gas: DefaultGas,
	}
	if !m.HasFee() {
		return "", &fee, nil
	}

	signWithAddr := m.ops[0].Account.Address
	if m.ops[0].Account.SubAccount != nil {
		return "", nil, fmt.Errorf("fee transfer from wrong subaccount (got: %s expected: nil)",
			m.ops[0].Account.SubAccount,
		)
	}
	feeAmount, err := readOasisCurrencyNeg(m.ops[0].Amount)
	if err != nil {
		return "", nil, fmt.Errorf("invalid fee transfer from amount: %w", err)
	}
	if feeGasRaw, ok := m.ops[0].Metadata[FeeGasKey]; ok {
		feeGasF64, ok := feeGasRaw.(float64)
		if !ok {
			return "", nil, fmt.Errorf("malformed fee transfer gas metadata")
		}
		fee.Gas = transaction.Gas(feeGasF64)
	}

	if m.ops[1].Account.SubAccount != nil {
		return "", nil, fmt.Errorf("fee transfer to wrong subaccount (got: %s expected: nil)",
			m.ops[1].Account.SubAccount,
		)
	}
	feeAmount2, err := readOasisCurrency(m.ops[1].Amount)
	if err != nil {
		return "", nil, fmt.Errorf("invalid fee transfer to amount: %w", err)
	}
	if feeAmount.Cmp(feeAmount2) != 0 {
		return "", nil, fmt.Errorf("fee transfer amounts differ between operations (from: %s to: %s)", feeAmount, feeAmount2)
	}
	fee.Amount = *feeAmount
	return signWithAddr, &fee, nil
}

func readCurrency(amount *types.Amount, currency *types.Currency, negative bool) (*quantity.Quantity, error) {
	// TODO: Is it up to us to check other fields?
	if amount.Currency.Symbol != currency.Symbol {
		return nil, fmt.Errorf("wrong currency")
	}
	bi := new(big.Int)
	if err := bi.UnmarshalText([]byte(amount.Value)); err != nil {
		return nil, fmt.Errorf("bi UnmarshalText Value: %w", err)
	}
	if negative {
		bi.Neg(bi)
	}
	q := quantity.NewQuantity()
	if err := q.FromBigInt(bi); err != nil {
		return nil, fmt.Errorf("q FromBigInt bi: %w", err)
	}
	return q, nil
}

func readOasisCurrency(amount *types.Amount) (*quantity.Quantity, error) {
	return readCurrency(amount, OasisCurrency, false)
}

func readOasisCurrencyNeg(amount *types.Amount) (*quantity.Quantity, error) {
	return readCurrency(amount, OasisCurrency, true)
}

// GetTransaction decodes an Oasis transaction from given Rosetta operations.
//
// The method also returns the signer address.
func (m *operationToTransactionMapper) GetTransaction() (string, *transaction.Transaction, error) {
	signWithAddr, fee, err := m.GetFee()
	if err != nil {
		return "", nil, fmt.Errorf("malformed fee operations: %w", err)
	}

	remainingOps := m.ops
	if m.HasFee() {
		remainingOps = remainingOps[2:]
	}

	var method transaction.MethodName
	var body cbor.RawMessage
	switch {
	case len(remainingOps) == 2 &&
		remainingOps[0].Type == OpTransfer &&
		remainingOps[0].Account.SubAccount == nil &&
		remainingOps[1].Type == OpTransfer &&
		remainingOps[1].Account.SubAccount == nil:
		// Staking transfer.
		method = staking.MethodTransfer

		if len(signWithAddr) == 0 {
			signWithAddr = remainingOps[0].Account.Address
		} else if remainingOps[0].Account.Address != signWithAddr {
			return "", nil, fmt.Errorf("transfer from doesn't match signer (from: %s signer: %s)",
				remainingOps[0].Account.Address,
				signWithAddr,
			)
		}
		amount, err := readOasisCurrencyNeg(remainingOps[0].Amount)
		if err != nil {
			return "", nil, fmt.Errorf("invalid transfer from amount: %w", err)
		}

		var to staking.Address
		if err = to.UnmarshalText([]byte(remainingOps[1].Account.Address)); err != nil {
			return "", nil, fmt.Errorf("invalid transfer to address (%s): %w", remainingOps[1].Account.Address, err)
		}
		amount2, err := readOasisCurrency(remainingOps[1].Amount)
		if err != nil {
			return "", nil, fmt.Errorf("invalid transfer to amount: %w", err)
		}
		if amount.Cmp(amount2) != 0 {
			return "", nil, fmt.Errorf("transfer amounts differ between operations (from: %s to: %s)", amount, amount2)
		}

		body = cbor.Marshal(staking.Transfer{
			To:     to,
			Amount: *amount,
		})
	case len(remainingOps) == 1 &&
		remainingOps[0].Type == OpBurn &&
		remainingOps[0].Account.SubAccount == nil:
		// Staking burn.
		method = staking.MethodBurn

		if len(signWithAddr) == 0 {
			signWithAddr = remainingOps[0].Account.Address
		} else if remainingOps[0].Account.Address != signWithAddr {
			return "", nil, fmt.Errorf("burn from doesn't match signer (from: %s signer: %s)",
				remainingOps[0].Account.Address,
				signWithAddr,
			)
		}
		amount, err := readOasisCurrencyNeg(remainingOps[0].Amount)
		if err != nil {
			return "", nil, fmt.Errorf("invalid burn from amount: %w", err)
		}

		body = cbor.Marshal(staking.Burn{
			Amount: *amount,
		})
	case len(remainingOps) == 2 &&
		remainingOps[0].Type == OpTransfer &&
		remainingOps[0].Account.SubAccount == nil &&
		remainingOps[1].Type == OpTransfer &&
		remainingOps[1].Account.SubAccount != nil &&
		remainingOps[1].Account.SubAccount.Address == SubAccountEscrow:
		// Staking add escrow.
		method = staking.MethodAddEscrow

		if len(signWithAddr) == 0 {
			signWithAddr = remainingOps[0].Account.Address
		} else if remainingOps[0].Account.Address != signWithAddr {
			return "", nil, fmt.Errorf("add escrow from doesn't match signer (from: %s signer: %s)",
				remainingOps[0].Account.Address,
				signWithAddr,
			)
		}
		amount, err := readOasisCurrencyNeg(remainingOps[0].Amount)
		if err != nil {
			return "", nil, fmt.Errorf("invalid add escrow from amount: %w", err)
		}

		var escrowAccount staking.Address
		if err = escrowAccount.UnmarshalText([]byte(remainingOps[1].Account.Address)); err != nil {
			return "", nil, fmt.Errorf("invalid add escrow to address (%s): %w", remainingOps[1].Account.Address, err)
		}
		amount2, err := readOasisCurrency(remainingOps[1].Amount)
		if err != nil {
			return "", nil, fmt.Errorf("invalid add escrow to amount: %w", err)
		}
		if amount.Cmp(amount2) != 0 {
			return "", nil, fmt.Errorf("add escrow amounts differ between operations (from: %s to: %s)", amount, amount2)
		}

		body = cbor.Marshal(staking.Escrow{
			Account: escrowAccount,
			Amount:  *amount,
		})
	case len(remainingOps) == 2 &&
		remainingOps[0].Type == OpReclaimEscrow &&
		remainingOps[1].Type == OpReclaimEscrow &&
		remainingOps[1].Account.SubAccount != nil &&
		remainingOps[1].Account.SubAccount.Address == SubAccountEscrow:
		// Staking reclaim escrow.
		method = staking.MethodReclaimEscrow

		if len(signWithAddr) == 0 {
			signWithAddr = remainingOps[0].Account.Address
		} else if remainingOps[0].Account.Address != signWithAddr {
			return "", nil, fmt.Errorf("reclaim escrow from doesn't match signer (from: %s signer: %s)",
				remainingOps[0].Account.Address,
				signWithAddr,
			)
		}
		if remainingOps[0].Amount != nil {
			return "", nil, fmt.Errorf("invalid reclaim escrow from amount (expected: nil): %w", err)
		}

		var escrowAccount staking.Address
		if err = escrowAccount.UnmarshalText([]byte(remainingOps[1].Account.Address)); err != nil {
			return "", nil, fmt.Errorf("invalid reclaim escrow address (%s): %w", remainingOps[1].Account.Address, err)
		}
		if remainingOps[1].Amount != nil {
			return "", nil, fmt.Errorf("invalid reclaim escrow to amount (expected: nil): %w", err)
		}
		sharesRaw, ok := remainingOps[1].Metadata[ReclaimEscrowSharesKey]
		if !ok {
			return "", nil, fmt.Errorf("reclaim escrow shares metadata not specified")
		}
		sharesStr, ok := sharesRaw.(string)
		if !ok {
			return "", nil, fmt.Errorf("malformed reclaim escrow shares metadata")
		}
		var shares quantity.Quantity
		if err = shares.UnmarshalText([]byte(sharesStr)); err != nil {
			return "", nil, fmt.Errorf("malformed reclaim escrow shares metadata (%s): %w", sharesStr, err)
		}

		body = cbor.Marshal(staking.ReclaimEscrow{
			Account: escrowAccount,
			Shares:  shares,
		})
	default:
		return "", nil, fmt.Errorf("not supported")
	}

	tx := &transaction.Transaction{
		// Nonce is omitted and must be filled in by caller.
		Fee:    fee,
		Method: method,
		Body:   body,
	}
	return signWithAddr, tx, nil
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

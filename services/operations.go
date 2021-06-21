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

var (
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
	if result != nil {
		d.decodeEvents(rosettaTx, result.Events)
	}
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
	//
	// Case 3:
	// * Result is not provided because the transaction has not yet been executed. In this case
	//   nothing has been emitted yet so we need to generate OK operations.
	if result == nil || !result.IsSuccess() {
		txSignerAddress := StringFromAddress(staking.NewAddress(sigTx.Signature.PublicKey))
		o2t := newOperationToTransactionMapper(rosettaTx.Operations)

		var status string
		switch {
		case result == nil:
			status = OpStatusOK
		default:
			status = OpStatusFailed
		}
		t2o := newTransactionToOperationMapper(&tx, txSignerAddress, status, rosettaTx.Operations)

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

func appendOp(
	ops []*types.Operation,
	kind, acct string,
	subacct *types.SubAccountIdentifier,
	amt string,
) []*types.Operation {
	opIndex := int64(len(ops))
	op := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: opIndex,
		},
		Type:   kind,
		Status: &OpStatusOK,
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
			{
				Index: opIndex - 1,
			},
		}
	}

	return append(ops, op)
}

func (d *transactionsDecoder) decodeStakingEvent(tx *types.Transaction, ev *staking.Event) {
	switch {
	case ev.Transfer != nil:
		tx.Operations = appendOp(
			tx.Operations,
			OpTransfer,
			StringFromAddress(ev.Transfer.From),
			nil,
			"-"+ev.Transfer.Amount.String(),
		)
		tx.Operations = appendOp(
			tx.Operations,
			OpTransfer,
			StringFromAddress(ev.Transfer.To),
			nil,
			ev.Transfer.Amount.String(),
		)
	case ev.Burn != nil:
		tx.Operations = appendOp(
			tx.Operations,
			OpBurn,
			StringFromAddress(ev.Burn.Owner),
			nil,
			"-"+ev.Burn.Amount.String(),
		)
	case ev.Escrow != nil:
		ee := ev.Escrow
		switch {
		case ee.Add != nil:
			// Owner's general account -> escrow account.
			tx.Operations = appendOp(
				tx.Operations,
				OpTransfer,
				StringFromAddress(ee.Add.Owner),
				nil,
				"-"+ee.Add.Amount.String(),
			)
			tx.Operations = appendOp(
				tx.Operations,
				OpTransfer,
				StringFromAddress(ee.Add.Escrow),
				&types.SubAccountIdentifier{Address: SubAccountEscrow},
				ee.Add.Amount.String(),
			)
		case ee.Take != nil:
			tx.Operations = appendOp(
				tx.Operations,
				OpTransfer,
				StringFromAddress(ee.Take.Owner),
				&types.SubAccountIdentifier{Address: SubAccountEscrow},
				"-"+ee.Take.Amount.String(),
			)
			tx.Operations = appendOp(
				tx.Operations,
				OpTransfer,
				StringFromAddress(staking.CommonPoolAddress),
				nil,
				ee.Take.Amount.String(),
			)
		case ee.Reclaim != nil:
			// Escrow account -> owner's general account.
			tx.Operations = appendOp(
				tx.Operations,
				OpTransfer,
				StringFromAddress(ee.Reclaim.Escrow),
				&types.SubAccountIdentifier{Address: SubAccountEscrow},
				"-"+ee.Reclaim.Amount.String(),
			)
			tx.Operations = appendOp(
				tx.Operations,
				OpTransfer,
				StringFromAddress(ee.Reclaim.Owner),
				nil,
				ee.Reclaim.Amount.String(),
			)
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

// getStakingTransfer decodes the Oasis staking transfer transaction from the
// given Rosetta operations.
func getStakingTransfer(ops []*types.Operation) (*staking.Transfer, error) {
	amount, err := readOasisCurrencyNeg(ops[0].Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid transfer from amount: %w", err)
	}

	var to staking.Address
	if err = to.UnmarshalText([]byte(ops[1].Account.Address)); err != nil {
		return nil, fmt.Errorf("invalid transfer's to address (%s): %w", ops[1].Account.Address, err)
	}
	amount2, err := readOasisCurrency(ops[1].Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid transfer's to amount: %w", err)
	}
	if amount.Cmp(amount2) != 0 {
		return nil, fmt.Errorf("transfer amounts differ between operations (from: %s to: %s)", amount, amount2)
	}

	xfer := staking.Transfer{
		To:     to,
		Amount: *amount,
	}
	return &xfer, nil
}

// getStakingBurn decodes the Oasis staking burn transaction from then given
// Rosetta operations.
func getStakingBurn(ops []*types.Operation) (*staking.Burn, error) {
	amount, err := readOasisCurrencyNeg(ops[0].Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid burn from amount: %w", err)
	}

	burn := staking.Burn{
		Amount: *amount,
	}
	return &burn, nil
}

// getStakingAddEscrow decodes the Oasis staking add escrow transaction from the
// given Rosetta operations.
func getStakingAddEscrow(ops []*types.Operation) (*staking.Escrow, error) {
	amount, err := readOasisCurrencyNeg(ops[0].Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid add escrow from amount: %w", err)
	}

	var escrowAccount staking.Address
	if err = escrowAccount.UnmarshalText([]byte(ops[1].Account.Address)); err != nil {
		return nil, fmt.Errorf("invalid add escrow to address (%s): %w", ops[1].Account.Address, err)
	}
	amount2, err := readOasisCurrency(ops[1].Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid add escrow to amount: %w", err)
	}
	if amount.Cmp(amount2) != 0 {
		return nil, fmt.Errorf("add escrow amounts differ between operations (from: %s to: %s)", amount, amount2)
	}

	escrow := staking.Escrow{
		Account: escrowAccount,
		Amount:  *amount,
	}
	return &escrow, nil
}

// getStakingReclaimEscrow decodes the Oasis staking reclaim escrow transaction
// from the given Rosetta operations.
func getStakingReclaimEscrow(ops []*types.Operation) (*staking.ReclaimEscrow, error) {
	if ops[0].Amount != nil {
		return nil, fmt.Errorf("invalid reclaim escrow's from amount (expected: nil): %s", ops[0].Amount.Value)
	}

	var escrowAccount staking.Address
	if err := escrowAccount.UnmarshalText([]byte(ops[1].Account.Address)); err != nil {
		return nil, fmt.Errorf("invalid reclaim escrow address (%s): %w", ops[1].Account.Address, err)
	}
	if ops[1].Amount != nil {
		return nil, fmt.Errorf("invalid reclaim escrow's to amount (expected: nil): %s", ops[1].Amount.Value)
	}
	sharesRaw, ok := ops[1].Metadata[ReclaimEscrowSharesKey]
	if !ok {
		return nil, fmt.Errorf("reclaim escrow shares metadata not specified")
	}
	sharesStr, ok := sharesRaw.(string)
	if !ok {
		return nil, fmt.Errorf("malformed reclaim escrow shares metadata")
	}
	var shares quantity.Quantity
	if err := shares.UnmarshalText([]byte(sharesStr)); err != nil {
		return nil, fmt.Errorf("malformed reclaim escrow shares metadata (%s): %w", sharesStr, err)
	}

	reclaim := staking.ReclaimEscrow{
		Account: escrowAccount,
		Shares:  shares,
	}
	return &reclaim, nil
}

// checkSigner ensures the operation's signer address matches the given signer
// address (if specified) and returns the operation's signer address.
func checkOpSignerAddress(op *types.Operation, signerAddr string) (string, error) {
	switch {
	case signerAddr == "":
		return op.Account.Address, nil
	case signerAddr != op.Account.Address:
		return "", fmt.Errorf("operation's address doesn't match signer's address (op: %s signer: %s)",
			op.Account.Address,
			signerAddr,
		)
	default:
		return signerAddr, nil
	}
}

// TransactionKind is the kind of Oasis transaction.
type TransactionKind int

const (
	KindUnknown              TransactionKind = 0
	KindStakingTransfer      TransactionKind = 1
	KindStakingBurn          TransactionKind = 2
	KindStakingAddEscrow     TransactionKind = 3
	KindStakingReclaimEscrow TransactionKind = 4
)

// decodeOpsToTransactionKind decodes the Oasis transaction kind from the given
// Rosetta operations.
func decodeOpsToTransactionKind(ops []*types.Operation) TransactionKind {
	switch {
	case len(ops) == 1:
		switch {
		case ops[0].Type == OpBurn &&
			ops[0].Account.SubAccount == nil:
			return KindStakingBurn
		default:
			return KindUnknown
		}
	case len(ops) == 2:
		switch {
		case ops[0].Type == OpTransfer &&
			ops[0].Account.SubAccount == nil &&
			ops[1].Type == OpTransfer:
			switch {
			case ops[1].Account.SubAccount == nil:
				return KindStakingTransfer
			case ops[1].Account.SubAccount != nil &&
				ops[1].Account.SubAccount.Address == SubAccountEscrow:
				return KindStakingAddEscrow
			default:
				return KindUnknown
			}
		case ops[0].Type == OpReclaimEscrow &&
			ops[1].Type == OpReclaimEscrow &&
			ops[1].Account.SubAccount != nil &&
			ops[1].Account.SubAccount.Address == SubAccountEscrow:
			return KindStakingReclaimEscrow
		default:
			return KindUnknown
		}
	default:
		return KindUnknown
	}
}

// GetTransaction decodes an Oasis transaction from given Rosetta operations.
//
// The method also returns the signer address.
func (m *operationToTransactionMapper) GetTransaction() (string, *transaction.Transaction, error) {
	signerAddr, fee, err := m.GetFee()
	if err != nil {
		return "", nil, fmt.Errorf("malformed fee operations: %w", err)
	}

	remainingOps := m.ops
	if m.HasFee() {
		remainingOps = remainingOps[2:]
	}

	decodedTxnKind := decodeOpsToTransactionKind(remainingOps)

	if decodedTxnKind != KindUnknown {
		signerAddr, err = checkOpSignerAddress(remainingOps[0], signerAddr)
		if err != nil {
			return "", nil, err
		}
	}

	var method transaction.MethodName
	var body cbor.RawMessage
	switch decodedTxnKind {
	case KindStakingTransfer:
		method = staking.MethodTransfer
		transfer, err2 := getStakingTransfer(remainingOps)
		if err2 != nil {
			return "", nil, err2
		}
		body = cbor.Marshal(transfer)
	case KindStakingBurn:
		method = staking.MethodBurn
		burn, err2 := getStakingBurn(remainingOps)
		if err2 != nil {
			return "", nil, err2
		}
		body = cbor.Marshal(burn)
	case KindStakingAddEscrow:
		method = staking.MethodAddEscrow
		addEscrow, err2 := getStakingAddEscrow(remainingOps)
		if err2 != nil {
			return "", nil, err2
		}
		body = cbor.Marshal(addEscrow)
	case KindStakingReclaimEscrow:
		method = staking.MethodReclaimEscrow
		reclaimEscrow, err2 := getStakingReclaimEscrow(remainingOps)
		if err2 != nil {
			return "", nil, err2
		}
		body = cbor.Marshal(reclaimEscrow)
	default:
		return "", nil, fmt.Errorf("not supported")
	}

	tx := &transaction.Transaction{
		// Nonce is omitted and must be filled in by caller.
		Fee:    fee,
		Method: method,
		Body:   body,
	}
	return signerAddr, tx, nil
}

func newOperationToTransactionMapper(ops []*types.Operation) *operationToTransactionMapper {
	return &operationToTransactionMapper{ops}
}

type transactionToOperationMapper struct {
	tx              *transaction.Transaction
	txSignerAddress string
	status          *string

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
				{
					Index: opIndex,
				},
			},
		},
	)
}

// emitTransferOps emits the required operations for the transfer transaction.
func (m *transactionToOperationMapper) emitTransferOps() error {
	var body staking.Transfer
	if err := cbor.Unmarshal(m.tx.Body, &body); err != nil {
		return fmt.Errorf("malformed body: %w", err)
	}

	opIndex := int64(len(m.ops))
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
				{
					Index: opIndex,
				},
			},
		},
	)

	return nil
}

// emitBurnOps emits the required operations for the burn transaction.
func (m *transactionToOperationMapper) emitBurnOps() error {
	var body staking.Burn
	if err := cbor.Unmarshal(m.tx.Body, &body); err != nil {
		return fmt.Errorf("malformed body: %w", err)
	}

	opIndex := int64(len(m.ops))
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

	return nil
}

// emitAddEscrowOps emits the required operations for the add escrow
// transaction.
func (m *transactionToOperationMapper) emitAddEscrowOps() error {
	var body staking.Escrow
	if err := cbor.Unmarshal(m.tx.Body, &body); err != nil {
		return fmt.Errorf("malformed body: %w", err)
	}

	opIndex := int64(len(m.ops))
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
				{
					Index: opIndex,
				},
			},
		},
	)

	return nil
}

// emitReclaimEscrowOps emits the required operations for the reclaim escrow
// transaction.
func (m *transactionToOperationMapper) emitReclaimEscrowOps() error {
	var body staking.ReclaimEscrow
	if err := cbor.Unmarshal(m.tx.Body, &body); err != nil {
		return fmt.Errorf("malformed body: %w", err)
	}

	opIndex := int64(len(m.ops))
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
				{
					Index: opIndex,
				},
			},
		},
	)

	return nil
}

// EmitTxOps emits the required transaction-specific operations.
func (m *transactionToOperationMapper) EmitTxOps() error {
	switch m.tx.Method {
	case staking.MethodTransfer:
		return m.emitTransferOps()
	case staking.MethodBurn:
		return m.emitBurnOps()
	case staking.MethodAddEscrow:
		return m.emitAddEscrowOps()
	case staking.MethodReclaimEscrow:
		return m.emitReclaimEscrowOps()
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
		status:          &status,
		ops:             ops,
	}
}

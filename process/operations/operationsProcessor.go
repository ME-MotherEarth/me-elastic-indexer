package operations

import (
	"encoding/hex"
	"strings"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/check"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	vmcommon "github.com/ME-MotherEarth/me-vm-common"
)

type operationsProcessor struct {
	importDBMode     bool
	shardCoordinator indexer.ShardCoordinator
}

// NewOperationsProcessor will create a new instance of operationsProcessor
func NewOperationsProcessor(importDBMode bool, shardCoordinator indexer.ShardCoordinator) (*operationsProcessor, error) {
	if check.IfNil(shardCoordinator) {
		return nil, indexer.ErrNilShardCoordinator
	}

	return &operationsProcessor{
		importDBMode:     importDBMode,
		shardCoordinator: shardCoordinator,
	}, nil
}

// ProcessTransactionsAndSCRs will prepare transactions and smart contract results to be indexed
func (op *operationsProcessor) ProcessTransactionsAndSCRs(
	txs []*data.Transaction,
	scrs []*data.ScResult,
) ([]*data.Transaction, []*data.ScResult) {
	newTxsSlice := make([]*data.Transaction, 0)
	newScrsSlice := make([]*data.ScResult, 0)

	for idx, tx := range txs {
		if !op.shouldIndex(txs[idx].ReceiverShard) {
			continue
		}

		copiedTx := *tx
		copiedTx.SmartContractResults = nil
		copiedTx.Type = string(transaction.TxTypeNormal)
		newTxsSlice = append(newTxsSlice, &copiedTx)
	}

	for idx := 0; idx < len(scrs); idx++ {
		if !op.shouldIndex(scrs[idx].ReceiverShard) {
			continue
		}

		copiedScr := *scrs[idx]
		copiedScr.Type = string(transaction.TxTypeUnsigned)

		setCanBeIgnoredField(&copiedScr)

		selfShard := op.shardCoordinator.SelfId()
		if selfShard == copiedScr.ReceiverShard {
			copiedScr.Status = transaction.TxStatusSuccess.String()
		} else {
			copiedScr.Status = transaction.TxStatusPending.String()
		}

		newScrsSlice = append(newScrsSlice, &copiedScr)
	}

	return newTxsSlice, newScrsSlice
}

func (op *operationsProcessor) shouldIndex(destinationShardID uint32) bool {
	if !op.importDBMode {
		return true
	}

	return op.shardCoordinator.SelfId() == destinationShardID
}

func setCanBeIgnoredField(scr *data.ScResult) {
	dataFieldStr := string(scr.Data)
	hasOkPrefix := strings.HasPrefix(dataFieldStr, data.AtSeparator+hex.EncodeToString([]byte(vmcommon.Ok.String())))
	isRefundForRelayed := scr.ReturnMessage == data.GasRefundForRelayerMessage && dataFieldStr == ""
	if hasOkPrefix || isRefundForRelayed {
		scr.CanBeIgnored = true
		return
	}

	isNFTTransferOrMultiTransfer := core.BuiltInFunctionMECTNFTTransfer == scr.Operation || core.BuiltInFunctionMultiMECTNFTTransfer == scr.Operation
	isSCAddr := core.IsSmartContractAddress(scr.SenderAddressBytes)
	if isNFTTransferOrMultiTransfer && !isSCAddr {
		scr.CanBeIgnored = true
		return
	}
}

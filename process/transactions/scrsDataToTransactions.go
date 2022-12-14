package transactions

import (
	"encoding/hex"
	"strings"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	vmcommon "github.com/ME-MotherEarth/me-vm-common"
)

const (
	minNumOfArgumentsNFTTransferORMultiTransfer = 4
)

type scrsDataToTransactions struct {
	retCodes        []string
	txFeeCalculator indexer.FeesProcessorHandler
}

func newScrsDataToTransactions(txFeeCalculator indexer.FeesProcessorHandler) *scrsDataToTransactions {
	return &scrsDataToTransactions{
		txFeeCalculator: txFeeCalculator,
		retCodes: []string{
			vmcommon.FunctionNotFound.String(),
			vmcommon.FunctionWrongSignature.String(),
			vmcommon.ContractNotFound.String(),
			vmcommon.UserError.String(),
			vmcommon.OutOfGas.String(),
			vmcommon.AccountCollision.String(),
			vmcommon.OutOfFunds.String(),
			vmcommon.CallStackOverFlow.String(),
			vmcommon.ContractInvalid.String(),
			vmcommon.ExecutionFailed.String(),
			vmcommon.UpgradeFailed.String(),
		},
	}
}

func (st *scrsDataToTransactions) attachSCRsToTransactionsAndReturnSCRsWithoutTx(txs map[string]*data.Transaction, scrs []*data.ScResult) []*data.ScResult {
	scrsWithoutTx := make([]*data.ScResult, 0)
	for _, scr := range scrs {
		decodedOriginalTxHash, err := hex.DecodeString(scr.OriginalTxHash)
		if err != nil {
			continue
		}

		tx, ok := txs[string(decodedOriginalTxHash)]
		if !ok {
			scrsWithoutTx = append(scrsWithoutTx, scr)
			continue
		}

		st.addScResultInfoIntoTx(scr, tx)
	}

	return scrsWithoutTx
}

func (st *scrsDataToTransactions) addScResultInfoIntoTx(dbScResult *data.ScResult, tx *data.Transaction) {
	tx.SmartContractResults = append(tx.SmartContractResults, dbScResult)
	isRelayedTxFirstSCR := isRelayedTx(tx) && len(tx.SmartContractResults) == 1
	if isRelayedTxFirstSCR {
		tx.GasUsed = tx.GasLimit
		fee := st.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
		tx.Fee = fee.String()
	}

	// ignore invalid transaction because status and gas fields was already set
	if tx.Status == transaction.TxStatusInvalid.String() {
		return
	}

	if isSCRForSenderWithRefund(dbScResult, tx) || isRefundForRelayed(dbScResult, tx) {
		refundValue := stringValueToBigInt(dbScResult.Value)
		gasUsed, fee := st.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
		tx.GasUsed = gasUsed
		tx.Fee = fee.String()
		tx.HadRefund = true
	}

	return
}

func (st *scrsDataToTransactions) processTransactionsAfterSCRsWereAttached(transactions map[string]*data.Transaction) {
	for _, tx := range transactions {
		if len(tx.SmartContractResults) == 0 {
			continue
		}

		st.fillTxWithSCRsFields(tx)
	}
}

func (st *scrsDataToTransactions) fillTxWithSCRsFields(tx *data.Transaction) {
	tx.HasSCR = true

	if isRelayedTx(tx) {
		return
	}

	// ignore invalid transaction because status and gas fields were already set
	if tx.Status == transaction.TxStatusInvalid.String() {
		return
	}

	if hasSuccessfulSCRs(tx) {
		return
	}

	tx.GasUsed = tx.GasLimit
	fee := st.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
	tx.Fee = fee.String()

	if hasCrossShardPendingTransfer(tx) {
		return
	}

	if st.hasSCRWithErrorCode(tx) {
		tx.Status = transaction.TxStatusFail.String()
	}
}

func (st *scrsDataToTransactions) hasSCRWithErrorCode(tx *data.Transaction) bool {
	for _, scr := range tx.SmartContractResults {
		for _, codeStr := range st.retCodes {
			if strings.Contains(string(scr.Data), hex.EncodeToString([]byte(codeStr))) ||
				scr.ReturnMessage == codeStr {
				return true
			}
		}
	}

	return false
}

func hasSuccessfulSCRs(tx *data.Transaction) bool {
	for _, scr := range tx.SmartContractResults {
		if isScResultSuccessful(scr.Data) {
			return true
		}
	}

	return false
}

func hasCrossShardPendingTransfer(tx *data.Transaction) bool {
	for _, scr := range tx.SmartContractResults {
		splitData := strings.Split(string(scr.Data), data.AtSeparator)
		if len(splitData) < 2 {
			continue
		}

		isMultiTransferOrNFTTransfer := splitData[0] == core.BuiltInFunctionMECTNFTTransfer || splitData[0] == core.BuiltInFunctionMultiMECTNFTTransfer
		if !isMultiTransferOrNFTTransfer {
			continue
		}

		if scr.SenderShard != scr.ReceiverShard {
			return true
		}
	}

	return false
}

func (st *scrsDataToTransactions) processSCRsWithoutTx(scrs []*data.ScResult) (map[string]string, map[string]*data.RefundData) {
	txHashStatus := make(map[string]string)
	txHashRefund := make(map[string]*data.RefundData)
	for _, scr := range scrs {
		if isSCRWithRefund(scr) {
			txHashRefund[scr.OriginalTxHash] = &data.RefundData{
				Value:    scr.Value,
				Receiver: scr.Receiver,
			}
		}

		if !isMECTNFTTransferWithUserError(string(scr.Data)) {
			continue
		}

		txHashStatus[scr.OriginalTxHash] = transaction.TxStatusFail.String()
	}

	return txHashStatus, txHashRefund
}

func isSCRWithRefund(scr *data.ScResult) bool {
	hasRefund := scr.Value != "0" && scr.Value != emptyString
	isSuccessful := isScResultSuccessful(scr.Data)
	isRefundForRelayTxSender := scr.ReturnMessage == data.GasRefundForRelayerMessage
	ok := isSuccessful || isRefundForRelayTxSender

	return ok && scr.OriginalTxHash != scr.PrevTxHash && hasRefund
}

func isMECTNFTTransferWithUserError(scrData string) bool {
	splitData := strings.Split(scrData, data.AtSeparator)
	isMultiTransferOrNFTTransfer := splitData[0] == core.BuiltInFunctionMECTNFTTransfer || splitData[0] == core.BuiltInFunctionMultiMECTNFTTransfer
	if !isMultiTransferOrNFTTransfer || len(splitData) < minNumOfArgumentsNFTTransferORMultiTransfer {
		return false
	}

	isUserErr := splitData[len(splitData)-1] == hex.EncodeToString([]byte(vmcommon.UserError.String()))

	return isUserErr
}

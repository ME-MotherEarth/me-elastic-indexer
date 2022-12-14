package transactions

import (
	"encoding/hex"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/check"
	coreData "github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-core/data/block"
	indexerArgs "github.com/ME-MotherEarth/me-core/data/indexer"
	"github.com/ME-MotherEarth/me-core/hashing"
	"github.com/ME-MotherEarth/me-core/marshal"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	logger "github.com/ME-MotherEarth/me-logger"
	datafield "github.com/ME-MotherEarth/me-vm-common/parsers/dataField"
)

var log = logger.GetOrCreate("indexer/process/transactions")

// ArgsTransactionProcessor holds all dependencies required by the txsDatabaseProcessor  in order to create
// new instances
type ArgsTransactionProcessor struct {
	AddressPubkeyConverter core.PubkeyConverter
	TxFeeCalculator        indexer.FeesProcessorHandler
	ShardCoordinator       indexer.ShardCoordinator
	Hasher                 hashing.Hasher
	Marshalizer            marshal.Marshalizer
	IsInImportMode         bool
}

type txsDatabaseProcessor struct {
	txFeeCalculator indexer.FeesProcessorHandler
	txBuilder       *dbTransactionBuilder
	txsGrouper      *txsGrouper
	scrsProc        *smartContractResultsProcessor
	scrsDataToTxs   *scrsDataToTransactions
}

// NewTransactionsProcessor will create a new instance of transactions database processor
func NewTransactionsProcessor(args *ArgsTransactionProcessor) (*txsDatabaseProcessor, error) {
	err := checkTxsProcessorArg(args)
	if err != nil {
		return nil, err
	}

	argsParser := &datafield.ArgsOperationDataFieldParser{
		AddressLength:    args.AddressPubkeyConverter.Len(),
		Marshalizer:      args.Marshalizer,
		ShardCoordinator: args.ShardCoordinator,
	}
	operationsDataParser, err := datafield.NewOperationDataFieldParser(argsParser)
	if err != nil {
		return nil, err
	}

	selfShardID := args.ShardCoordinator.SelfId()
	txBuilder := newTransactionDBBuilder(args.AddressPubkeyConverter, args.ShardCoordinator, args.TxFeeCalculator, operationsDataParser)
	txsDBGrouper := newTxsGrouper(txBuilder, args.IsInImportMode, selfShardID, args.Hasher, args.Marshalizer)
	scrProc := newSmartContractResultsProcessor(args.AddressPubkeyConverter, args.ShardCoordinator, args.Marshalizer, args.Hasher, operationsDataParser)
	scrsDataToTxs := newScrsDataToTransactions(args.TxFeeCalculator)

	if args.IsInImportMode {
		log.Warn("the node is in import mode! Cross shard transactions and rewards where destination shard is " +
			"not the current node's shard won't be indexed in Elastic Search")
	}

	return &txsDatabaseProcessor{
		txFeeCalculator: args.TxFeeCalculator,
		txBuilder:       txBuilder,
		txsGrouper:      txsDBGrouper,
		scrsProc:        scrProc,
		scrsDataToTxs:   scrsDataToTxs,
	}, nil
}

// PrepareTransactionsForDatabase will prepare transactions for database
func (tdp *txsDatabaseProcessor) PrepareTransactionsForDatabase(
	body *block.Body,
	header coreData.HeaderHandler,
	pool *indexerArgs.Pool,
) *data.PreparedResults {
	err := checkPrepareTransactionForDatabaseArguments(body, header, pool)
	if err != nil {
		log.Warn("checkPrepareTransactionForDatabaseArguments", "error", err)

		return &data.PreparedResults{
			Transactions: []*data.Transaction{},
			ScResults:    []*data.ScResult{},
			Receipts:     []*data.Receipt{},
			AlteredAccts: data.NewAlteredAccounts(),
		}
	}

	alteredAccounts := data.NewAlteredAccounts()
	normalTxs := make(map[string]*data.Transaction)
	rewardsTxs := make(map[string]*data.Transaction)

	for mbIndex, mb := range body.MiniBlocks {
		switch mb.Type {
		case block.TxBlock:
			if shouldIgnoreProcessedMBScheduled(header, mbIndex) {
				continue
			}

			txs, errGroup := tdp.txsGrouper.groupNormalTxs(mbIndex, mb, header, pool.Txs, alteredAccounts)
			if errGroup != nil {
				log.Warn("txsDatabaseProcessor.groupNormalTxs", "error", errGroup)
				continue
			}
			mergeTxsMaps(normalTxs, txs)
		case block.RewardsBlock:
			txs, errGroup := tdp.txsGrouper.groupRewardsTxs(mbIndex, mb, header, pool.Rewards, alteredAccounts)
			if errGroup != nil {
				log.Warn("txsDatabaseProcessor.groupRewardsTxs", "error", errGroup)
				continue
			}
			mergeTxsMaps(rewardsTxs, txs)
		case block.InvalidBlock:
			txs, errGroup := tdp.txsGrouper.groupInvalidTxs(mbIndex, mb, header, pool.Invalid, alteredAccounts)
			if errGroup != nil {
				log.Warn("txsDatabaseProcessor.groupInvalidTxs", "error", errGroup)
				continue
			}
			mergeTxsMaps(normalTxs, txs)
		default:
			continue
		}
	}

	normalTxs = tdp.setTransactionSearchOrder(normalTxs)
	dbReceipts := tdp.txsGrouper.groupReceipts(header, pool.Receipts)
	dbSCResults := tdp.scrsProc.processSCRs(body, header, pool.Scrs)

	tdp.scrsProc.addScrsReceiverToAlteredAccounts(alteredAccounts, dbSCResults)

	srcsNoTxInCurrentShard := tdp.scrsDataToTxs.attachSCRsToTransactionsAndReturnSCRsWithoutTx(normalTxs, dbSCResults)
	tdp.scrsDataToTxs.processTransactionsAfterSCRsWereAttached(normalTxs)
	txHashStatus, txHashRefund := tdp.scrsDataToTxs.processSCRsWithoutTx(srcsNoTxInCurrentShard)

	sliceNormalTxs := convertMapTxsToSlice(normalTxs)
	sliceRewardsTxs := convertMapTxsToSlice(rewardsTxs)
	txsSlice := append(sliceNormalTxs, sliceRewardsTxs...)

	return &data.PreparedResults{
		Transactions: txsSlice,
		ScResults:    dbSCResults,
		Receipts:     dbReceipts,
		AlteredAccts: alteredAccounts,
		TxHashStatus: txHashStatus,
		TxHashRefund: txHashRefund,
	}
}

func (tdp *txsDatabaseProcessor) setTransactionSearchOrder(transactions map[string]*data.Transaction) map[string]*data.Transaction {
	currentOrder := uint32(0)
	for _, tx := range transactions {
		tx.SearchOrder = currentOrder
		currentOrder++
	}

	return transactions
}

// GetHexEncodedHashesForRemove will return hex encoded transaction hashes and smart contract result hashes from body
func (tdp *txsDatabaseProcessor) GetHexEncodedHashesForRemove(header coreData.HeaderHandler, body *block.Body) ([]string, []string) {
	if body == nil || check.IfNil(header) || len(header.GetMiniBlockHeadersHashes()) == 0 {
		return nil, nil
	}

	selfShardID := header.GetShardID()
	encodedTxsHashes := make([]string, 0)
	encodedScrsHashes := make([]string, 0)
	for _, miniblock := range body.MiniBlocks {
		shouldIgnore := isCrossShardAtSourceNormalTx(selfShardID, miniblock)
		if shouldIgnore {
			// ignore cross-shard miniblocks at source with normal txs
			continue
		}

		if tdp.txsGrouper.isInImportMode {
			// do not delete transactions on import DB
			continue
		}

		txsHashesFromMiniblock := getTxsHashesFromMiniblockHexEncoded(miniblock)
		if miniblock.Type == block.SmartContractResultBlock {
			encodedScrsHashes = append(encodedScrsHashes, txsHashesFromMiniblock...)
			continue
		}
		encodedTxsHashes = append(encodedTxsHashes, txsHashesFromMiniblock...)
	}

	return encodedTxsHashes, encodedScrsHashes
}

func isCrossShardAtSourceNormalTx(selfShardID uint32, miniblock *block.MiniBlock) bool {
	isCrossShard := miniblock.SenderShardID != miniblock.ReceiverShardID
	isAtSource := miniblock.SenderShardID == selfShardID
	txBlock := miniblock.Type == block.TxBlock

	return isCrossShard && isAtSource && txBlock
}

func shouldIgnoreProcessedMBScheduled(header coreData.HeaderHandler, mbIndex int) bool {
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniblockHeaders) <= mbIndex {
		return false
	}

	processingType := miniblockHeaders[mbIndex].GetProcessingType()

	return processingType == int32(block.Processed)
}

func getTxsHashesFromMiniblockHexEncoded(miniBlock *block.MiniBlock) []string {
	encodedTxsHashes := make([]string, 0)
	for _, txHash := range miniBlock.TxHashes {
		encodedTxsHashes = append(encodedTxsHashes, hex.EncodeToString(txHash))
	}

	return encodedTxsHashes
}

func mergeTxsMaps(dst, src map[string]*data.Transaction) {
	for key, value := range src {
		dst[key] = value
	}
}

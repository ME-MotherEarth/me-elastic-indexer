package indexer

import (
	"math/big"

	coreData "github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/data/indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/workItems"
	vmcommon "github.com/ME-MotherEarth/me-vm-common"
)

// DispatcherHandler defines the interface for the dispatcher that will manage when items are saved in elasticsearch database
type DispatcherHandler interface {
	StartIndexData()
	Close() error
	Add(item workItems.WorkItemHandler)
	IsInterfaceNil() bool
}

// ElasticProcessor defines the interface for the elastic search indexer
type ElasticProcessor interface {
	SaveHeader(
		headerHash []byte,
		header coreData.HeaderHandler,
		signersIndexes []uint64,
		body *block.Body,
		notarizedHeadersHashes []string,
		gasConsumptionData indexer.HeaderGasConsumption,
		txsSize int,
	) error
	RemoveHeader(header coreData.HeaderHandler) error
	RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error
	RemoveAccountsMECT(headerTimestamp uint64) error
	SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error
	SaveTransactions(body *block.Body, header coreData.HeaderHandler, pool *indexer.Pool) error
	SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error
	SaveRoundsInfo(infos []*data.RoundInfo) error
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
	SaveAccounts(blockTimestamp uint64, accounts []*data.Account) error
	IsInterfaceNil() bool
}

// FeesProcessorHandler defines the interface for the transaction fees processor
type FeesProcessorHandler interface {
	ComputeGasUsedAndFeeBasedOnRefundValue(tx coreData.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsed(tx coreData.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimit(tx coreData.TransactionWithFeeHandler) uint64
	IsInterfaceNil() bool
}

// ShardCoordinator defines what a shard state coordinator should hold
type ShardCoordinator interface {
	NumberOfShards() uint32
	ComputeId(address []byte) uint32
	SelfId() uint32
	SameShard(firstAddress, secondAddress []byte) bool
	CommunicationIdentifier(destShardID uint32) string
	IsInterfaceNil() bool
}

// Indexer is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Indexer interface {
	SaveBlock(args *indexer.ArgsSaveBlockData) error
	RevertIndexedBlock(header coreData.HeaderHandler, body coreData.BodyHandler) error
	SaveRoundsInfo(roundsInfos []*indexer.RoundInfo) error
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) error
	SaveValidatorsRating(indexID string, infoRating []*indexer.ValidatorRatingInfo) error
	SaveAccounts(blockTimestamp uint64, acc []coreData.UserAccountHandler) error
	FinalizedBlock(headerHash []byte) error
	Close() error
	IsInterfaceNil() bool
	IsNilIndexer() bool
}

type AccountsAdapter interface {
	LoadAccount(address []byte) (vmcommon.AccountHandler, error)
	IsInterfaceNil() bool
}

// BalanceConverter defines what a balance converter should be able to do
type BalanceConverter interface {
	ComputeBalanceAsFloat(balance *big.Int) float64
	ComputeMECTBalanceAsFloat(balance *big.Int) float64
	IsInterfaceNil() bool
}

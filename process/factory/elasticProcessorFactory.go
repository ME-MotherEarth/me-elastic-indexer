package factory

import (
	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/hashing"
	"github.com/ME-MotherEarth/me-core/marshal"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/converters"
	processIndexer "github.com/ME-MotherEarth/me-elastic-indexer/process"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/accounts"
	blockProc "github.com/ME-MotherEarth/me-elastic-indexer/process/block"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/logsevents"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/miniblocks"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/operations"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/statistics"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/templatesAndPolicies"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/transactions"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/validators"
)

// ArgElasticProcessorFactory is struct that is used to store all components that are needed to create an elastic processor factory
type ArgElasticProcessorFactory struct {
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	DBClient                 processIndexer.DatabaseClientHandler
	AccountsDB               indexer.AccountsAdapter
	ShardCoordinator         indexer.ShardCoordinator
	TransactionFeeCalculator indexer.FeesProcessorHandler
	EnabledIndexes           []string
	Denomination             int
	BulkRequestMaxSize       int
	IsInImportDBMode         bool
	UseKibana                bool
}

// CreateElasticProcessor will create a new instance of ElasticProcessor
func CreateElasticProcessor(arguments ArgElasticProcessorFactory) (indexer.ElasticProcessor, error) {
	templatesAndPoliciesReader := templatesAndPolicies.CreateTemplatesAndPoliciesReader(arguments.UseKibana)
	indexTemplates, indexPolicies, err := templatesAndPoliciesReader.GetElasticTemplatesAndPolicies()
	if err != nil {
		return nil, err
	}

	enabledIndexesMap := make(map[string]struct{})
	for _, index := range arguments.EnabledIndexes {
		enabledIndexesMap[index] = struct{}{}
	}
	if len(enabledIndexesMap) == 0 {
		return nil, indexer.ErrEmptyEnabledIndexes
	}

	balanceConverter, err := converters.NewBalanceConverter(arguments.Denomination)
	if err != nil {
		return nil, err
	}

	accountsProc, err := accounts.NewAccountsProcessor(
		arguments.Marshalizer,
		arguments.AddressPubkeyConverter,
		arguments.AccountsDB,
		balanceConverter,
		arguments.ShardCoordinator.SelfId(),
	)
	if err != nil {
		return nil, err
	}

	blockProcHandler, err := blockProc.NewBlockProcessor(arguments.Hasher, arguments.Marshalizer)
	if err != nil {
		return nil, err
	}

	miniblocksProc, err := miniblocks.NewMiniblocksProcessor(arguments.ShardCoordinator.SelfId(), arguments.Hasher, arguments.Marshalizer, arguments.IsInImportDBMode)
	if err != nil {
		return nil, err
	}
	validatorsProc, err := validators.NewValidatorsProcessor(arguments.ValidatorPubkeyConverter, arguments.BulkRequestMaxSize)
	if err != nil {
		return nil, err
	}

	generalInfoProc := statistics.NewStatisticsProcessor()

	argsTxsProc := &transactions.ArgsTransactionProcessor{
		AddressPubkeyConverter: arguments.AddressPubkeyConverter,
		TxFeeCalculator:        arguments.TransactionFeeCalculator,
		ShardCoordinator:       arguments.ShardCoordinator,
		Hasher:                 arguments.Hasher,
		Marshalizer:            arguments.Marshalizer,
		IsInImportMode:         arguments.IsInImportDBMode,
	}
	txsProc, err := transactions.NewTransactionsProcessor(argsTxsProc)
	if err != nil {
		return nil, err
	}

	argsLogsAndEventsProc := &logsevents.ArgsLogsAndEventsProcessor{
		ShardCoordinator: arguments.ShardCoordinator,
		PubKeyConverter:  arguments.AddressPubkeyConverter,
		Marshalizer:      arguments.Marshalizer,
		BalanceConverter: balanceConverter,
		Hasher:           arguments.Hasher,
		TxFeeCalculator:  arguments.TransactionFeeCalculator,
	}
	logsAndEventsProc, err := logsevents.NewLogsAndEventsProcessor(argsLogsAndEventsProc)
	if err != nil {
		return nil, err
	}

	operationsProc, err := operations.NewOperationsProcessor(arguments.IsInImportDBMode, arguments.ShardCoordinator)
	if err != nil {
		return nil, err
	}

	args := &processIndexer.ArgElasticProcessor{
		BulkRequestMaxSize: arguments.BulkRequestMaxSize,
		TransactionsProc:   txsProc,
		AccountsProc:       accountsProc,
		BlockProc:          blockProcHandler,
		MiniblocksProc:     miniblocksProc,
		ValidatorsProc:     validatorsProc,
		StatisticsProc:     generalInfoProc,
		LogsAndEventsProc:  logsAndEventsProc,
		DBClient:           arguments.DBClient,
		EnabledIndexes:     enabledIndexesMap,
		UseKibana:          arguments.UseKibana,
		IndexTemplates:     indexTemplates,
		IndexPolicies:      indexPolicies,
		SelfShardID:        arguments.ShardCoordinator.SelfId(),
		OperationsProc:     operationsProc,
	}

	return processIndexer.NewElasticProcessor(args)
}

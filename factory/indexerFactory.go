package factory

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/check"
	"github.com/ME-MotherEarth/me-core/hashing"
	"github.com/ME-MotherEarth/me-core/marshal"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/client"
	"github.com/ME-MotherEarth/me-elastic-indexer/client/logging"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/factory"
	logger "github.com/ME-MotherEarth/me-logger"
	"github.com/elastic/go-elasticsearch/v7"
)

var log = logger.GetOrCreate("indexer/factory")

// ArgsIndexerFactory holds all dependencies required by the data indexer factory in order to create
// new instances
type ArgsIndexerFactory struct {
	Enabled                  bool
	UseKibana                bool
	IsInImportDBMode         bool
	IndexerCacheSize         int
	Denomination             int
	BulkRequestMaxSize       int
	Url                      string
	UserName                 string
	Password                 string
	TemplatesPath            string
	EnabledIndexes           []string
	ShardCoordinator         indexer.ShardCoordinator
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	AccountsDB               indexer.AccountsAdapter
	TransactionFeeCalculator indexer.FeesProcessorHandler
}

// NewIndexer will create a new instance of Indexer
func NewIndexer(args *ArgsIndexerFactory) (indexer.Indexer, error) {
	err := checkDataIndexerParams(args)
	if err != nil {
		return nil, err
	}

	if !args.Enabled {
		return indexer.NewNilIndexer(), nil
	}

	elasticProcessor, err := createElasticProcessor(args)
	if err != nil {
		return nil, err
	}

	dispatcher, err := indexer.NewDataDispatcher(args.IndexerCacheSize)
	if err != nil {
		return nil, err
	}

	dispatcher.StartIndexData()

	arguments := indexer.ArgDataIndexer{
		Marshalizer:      args.Marshalizer,
		ShardCoordinator: args.ShardCoordinator,
		ElasticProcessor: elasticProcessor,
		DataDispatcher:   dispatcher,
	}

	return indexer.NewDataIndexer(arguments)
}

func retryBackOff(attempt int) time.Duration {
	d := time.Duration(math.Exp2(float64(attempt))) * time.Second
	log.Debug("elastic: retry backoff", "attempt", attempt, "sleep duration", d)

	return d
}

func createElasticProcessor(args *ArgsIndexerFactory) (indexer.ElasticProcessor, error) {
	databaseClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses:     []string{args.Url},
		Username:      args.UserName,
		Password:      args.Password,
		Logger:        &logging.CustomLogger{},
		RetryOnStatus: []int{http.StatusConflict},
		RetryBackoff:  retryBackOff,
	})
	if err != nil {
		return nil, err
	}

	argsElasticProcFac := factory.ArgElasticProcessorFactory{
		Marshalizer:              args.Marshalizer,
		Hasher:                   args.Hasher,
		AddressPubkeyConverter:   args.AddressPubkeyConverter,
		ValidatorPubkeyConverter: args.ValidatorPubkeyConverter,
		UseKibana:                args.UseKibana,
		DBClient:                 databaseClient,
		AccountsDB:               args.AccountsDB,
		Denomination:             args.Denomination,
		TransactionFeeCalculator: args.TransactionFeeCalculator,
		IsInImportDBMode:         args.IsInImportDBMode,
		ShardCoordinator:         args.ShardCoordinator,
		EnabledIndexes:           args.EnabledIndexes,
		BulkRequestMaxSize:       args.BulkRequestMaxSize,
	}

	return factory.CreateElasticProcessor(argsElasticProcFac)
}

func checkDataIndexerParams(arguments *ArgsIndexerFactory) error {
	if arguments.IndexerCacheSize < 0 {
		return indexer.ErrNegativeCacheSize
	}
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return fmt.Errorf("%w when setting AddressPubkeyConverter in indexer", indexer.ErrNilPubkeyConverter)
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return fmt.Errorf("%w when setting ValidatorPubkeyConverter in indexer", indexer.ErrNilPubkeyConverter)
	}
	if arguments.Url == "" {
		return indexer.ErrNilUrl
	}
	if check.IfNil(arguments.Marshalizer) {
		return indexer.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return indexer.ErrNilHasher
	}
	if check.IfNil(arguments.TransactionFeeCalculator) {
		return indexer.ErrNilTransactionFeeCalculator
	}
	if check.IfNil(arguments.AccountsDB) {
		return indexer.ErrNilAccountsDB
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return indexer.ErrNilShardCoordinator
	}

	return nil
}

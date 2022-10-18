package integrationtests

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/client"
	"github.com/ME-MotherEarth/me-elastic-indexer/client/logging"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/ME-MotherEarth/me-elastic-indexer/process"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/factory"
	logger "github.com/ME-MotherEarth/me-logger"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/require"
)

const esURL = "http://localhost:9200"

func setLogLevelDebug() {
	_ = logger.SetLogLevel("process:DEBUG")
}

func createESClient(url string) (process.DatabaseClientHandler, error) {
	return client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{url},
		Logger:    &logging.CustomLogger{},
	})
}

// CreateElasticProcessor -
func CreateElasticProcessor(
	esClient process.DatabaseClientHandler,
	accountsDB indexer.AccountsAdapter,
	shardCoordinator indexer.ShardCoordinator,
	feeProcessor indexer.FeesProcessorHandler,
) (indexer.ElasticProcessor, error) {
	args := factory.ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		DBClient:                 esClient,
		AccountsDB:               accountsDB,
		ShardCoordinator:         shardCoordinator,
		TransactionFeeCalculator: feeProcessor,
		EnabledIndexes: []string{indexer.TransactionsIndex, indexer.LogsIndex, indexer.AccountsMECTIndex, indexer.ScResultsIndex,
			indexer.ReceiptsIndex, indexer.BlockIndex, indexer.AccountsIndex, indexer.TokensIndex, indexer.TagsIndex, indexer.CollectionsIndex,
			indexer.OperationsIndex},
		Denomination:     18,
		IsInImportDBMode: false,
	}

	return factory.CreateElasticProcessor(args)
}

func compareTxs(t *testing.T, expected []byte, actual []byte) {
	expectedTx := &data.Transaction{}
	err := json.Unmarshal(expected, expectedTx)
	require.Nil(t, err)

	actualTx := &data.Transaction{}
	err = json.Unmarshal(actual, actualTx)
	require.Nil(t, err)

	require.Equal(t, expectedTx, actualTx)
}

func readExpectedResult(path string) string {
	jsonFile, _ := os.Open(path)
	byteValue, _ := ioutil.ReadAll(jsonFile)

	return string(byteValue)
}

func getElementFromSlice(path string, index int) string {
	fileBytes := readExpectedResult(path)
	slice := make([]map[string]interface{}, 0)
	_ = json.Unmarshal([]byte(fileBytes), &slice)
	res, _ := json.Marshal(slice[index]["_source"])

	return string(res)
}

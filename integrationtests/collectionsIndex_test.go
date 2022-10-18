//go:build integrationtests

package integrationtests

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ME-MotherEarth/me-core/core"
	coreData "github.com/ME-MotherEarth/me-core/data"
	dataBlock "github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/data/indexer"
	"github.com/ME-MotherEarth/me-core/data/mect"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	indexerdata "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	vmcommon "github.com/ME-MotherEarth/me-vm-common"
	"github.com/stretchr/testify/require"
)

func TestCollectionsIndexInsertAndDelete(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	// ################ ISSUE NON FUNGIBLE TOKEN ##########################
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("SSSS-dddd"), []byte("SEMI-semi"), []byte("SSS"), []byte(core.SemiFungibleMECT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	// ################ CREATE SEMI FUNGIBLE TOKEN 1 ##########################
	shardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: 0,
	}

	mectToken := &mect.MECToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
		TokenMetaData: &mect.MetaData{
			Creator: []byte("creator"),
		},
	}

	addr := "aaaabbbbcccccccc"
	mockAccount := &mock.UserAccountStub{
		RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
			return json.Marshal(mectToken)
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}
	accounts = &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	esProc, err = CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
	}

	mectData := &mect.MECToken{
		TokenMetaData: &mect.MetaData{
			Creator: []byte("creator"),
		},
	}
	mectDataBytes, _ := json.Marshal(mectData)

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreate),
							Topics:     [][]byte{[]byte("SSSS-dddd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), mectDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)
	ids := []string{"61616161626262626363636363636363"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.CollectionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/collectionsIndex/collections-1.json"), string(genericResponse.Docs[0].Source))

	// ################ CREATE SEMI FUNGIBLE TOKEN 2 ##########################
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreate),
							Topics:     [][]byte{[]byte("SSSS-dddd"), big.NewInt(22).Bytes(), big.NewInt(1).Bytes(), mectDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)
	ids = []string{"61616161626262626363636363636363"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.CollectionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/collectionsIndex/collections-2.json"), string(genericResponse.Docs[0].Source))

	// ################ TRANSFER SEMI FUNGIBLE TOKEN 2 ##########################
	mectToken = &mect.MECToken{
		Value:      big.NewInt(0),
		Properties: []byte("ok"),
		TokenMetaData: &mect.MetaData{
			Creator: []byte("creator"),
		},
	}

	addr = "aaaabbbbcccccccc"
	mockAccount = &mock.UserAccountStub{
		RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
			return json.Marshal(mectToken)
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}
	accounts = &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	esProc, err = CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte(addr),
							Identifier: []byte(core.BuiltInFunctionMECTNFTTransfer),
							Topics:     [][]byte{[]byte("SSSS-dddd"), big.NewInt(22).Bytes(), big.NewInt(1).Bytes(), []byte("746573742d616464726573732d62616c616e63652d31")},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)
	ids = []string{"61616161626262626363636363636363"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.CollectionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/collectionsIndex/collections-1.json"), string(genericResponse.Docs[0].Source))
}

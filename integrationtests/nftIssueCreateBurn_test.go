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
	"github.com/stretchr/testify/require"
)

func TestIssueNFTCreateAndBurn(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}

	// ################ ISSUE NON FUNGIBLE TOKEN ##########################
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}

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
							Identifier: []byte("issueNonFungible"),
							Topics:     [][]byte{[]byte("NON-abcd"), []byte("NON-token"), []byte("NON"), []byte(core.NonFungibleMECT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"NON-abcd"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/nftIssueCreateBurn/non-fungible-after-issue.json"), string(genericResponse.Docs[0].Source))

	// ################ CREATE NON FUNGIBLE TOKEN ##########################
	shardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: 0,
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
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreate),
							Topics:     [][]byte{[]byte("NON-abcd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), mectDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"NON-abcd-02"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/nftIssueCreateBurn/non-fungible-after-create.json"), string(genericResponse.Docs[0].Source))

	// ################ BURN NON FUNGIBLE TOKEN ##########################

	header = &dataBlock.Header{
		Round:     52,
		TimeStamp: 5666,
	}

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTBurn),
							Topics:     [][]byte{[]byte("NON-abcd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), []byte("adr")},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"NON-abcd-02"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)
}

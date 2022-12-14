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

func TestNFTUpdateMetadata(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	mectCreateData := &mect.MECToken{
		TokenMetaData: &mect.MetaData{
			URIs: [][]byte{[]byte("uri"), []byte("uri")},
		},
	}
	marshalizedCreate, _ := json.Marshal(mectCreateData)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	body := &dataBlock.Body{}

	// CREATE NFT data
	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreate),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"NFT-abcd-0e"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token.json"), string(genericResponse.Docs[0].Source))

	// Add URIS 1
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTAddURI),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri1"), []byte("uri2")},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	// Add URIS 2 --- results should be the same
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTAddURI),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri1"), []byte("uri2")},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	// Update attributes 1
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-add-uris.json"), string(genericResponse.Docs[0].Source))

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTUpdateAttributes),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("tags:test,free,fun;description:This is a test description for an awesome nft;metadata:metadata-test")},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update-attributes.json"), string(genericResponse.Docs[0].Source))

	// Update attributes 2

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTUpdateAttributes),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("something")},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update-attributes-second.json"), string(genericResponse.Docs[0].Source))
}

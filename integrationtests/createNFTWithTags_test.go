//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
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

func TestCreateNFTWithTags(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	mectToken := &mect.MECToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
		TokenMetaData: &mect.MetaData{
			Creator:    []byte("creator"),
			Attributes: []byte("tags:hello,something,do,music,art,gallery;metadata:QmZ2QqaGq4bqsEzs5JLTjRmmvR2GAR4qXJZBN8ibfDdaud"),
		},
	}

	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: 0,
	}

	header := &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
	}

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{
		RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
			return json.Marshal(mectToken)
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}
	accounts := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	mectDataBytes, _ := json.Marshal(mectToken)

	// CREATE A FIRST NFT WITH THE TAGS
	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreate),
							Topics:     [][]byte{[]byte("DESK-abcd"), big.NewInt(1).Bytes(), big.NewInt(1).Bytes(), mectDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	body := &dataBlock.Body{}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"6161616162626262-DESK-abcd-01"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/createNFTWithTags/accounts-mect-address-balance.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"bXVzaWM=", "aGVsbG8=", "Z2FsbGVyeQ==", "ZG8=", "YXJ0", "c29tZXRoaW5n"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TagsIndex, true, genericResponse)
	require.Nil(t, err)

	tagsChecked := 0
	for idx, id := range ids {
		expectedDoc := getElementFromSlice("./testdata/createNFTWithTags/tags1.json", idx)
		for _, doc := range genericResponse.Docs {
			if doc.ID == id {
				require.JSONEq(t, expectedDoc, string(doc.Source))
				tagsChecked++
			}
		}
	}
	require.Equal(t, len(ids), tagsChecked)

	// CREATE A SECOND NFT WITH THE SAME TAGS
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreate),
							Topics:     [][]byte{[]byte("DESK-abcd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), mectDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	body = &dataBlock.Body{}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TagsIndex, true, genericResponse)
	require.Nil(t, err)

	tagsChecked = 0
	for idx, id := range ids {
		expectedDoc := getElementFromSlice("./testdata/createNFTWithTags/tags2.json", idx)
		for _, doc := range genericResponse.Docs {
			if doc.ID == id {
				require.JSONEq(t, expectedDoc, string(doc.Source))
				tagsChecked++
			}
		}
	}
	require.Equal(t, len(ids), tagsChecked)

	// CREATE A 3RD NFT WITH THE SPECIAL TAGS
	hexEncodedAttributes := "746167733a5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c2c3c3c3c3e3e3e2626262626262626262626262626262c272727273b6d657461646174613a516d533757525566464464516458654c513637516942394a33663746654d69343554526d6f79415741563568345a"
	attributes, _ := hex.DecodeString(hexEncodedAttributes)

	mectToken = &mect.MECToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
		TokenMetaData: &mect.MetaData{
			Creator:    []byte("creator"),
			Attributes: attributes,
		},
	}
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
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreate),
							Topics:     [][]byte{[]byte("DESK-abcd"), big.NewInt(3).Bytes(), big.NewInt(1).Bytes(), mectDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	body = &dataBlock.Body{}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = append(ids, "XFxcXFxcXFxcXFxcXFxcXFxcXA==", "JycnJw==", "PDw8Pj4+JiYmJiYmJiYmJiYmJiYm")
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TagsIndex, true, genericResponse)
	require.Nil(t, err)

	tagsChecked = 0
	for idx, id := range ids {
		expectedDoc := getElementFromSlice("./testdata/createNFTWithTags/tags3.json", idx)
		for _, doc := range genericResponse.Docs {
			if doc.ID == id {
				require.JSONEq(t, expectedDoc, string(doc.Source))
				tagsChecked++
			}
		}
	}
	require.Equal(t, len(ids), tagsChecked)
}

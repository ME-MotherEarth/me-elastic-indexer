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

func TestAccountsMECTDeleteOnRollback(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	feeComputer := &mock.EconomicsHandlerMock{}

	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: 1,
		ComputeIdCalled: func(address []byte) uint32 {
			return 1
		},
	}

	mectToken := &mect.MECToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
		TokenMetaData: &mect.MetaData{
			Creator: []byte("creator"),
		},
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

	// CREATE SEMI-FUNGIBLE TOKEN
	mectDataBytes, _ := json.Marshal(mectToken)
	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreate),
							Topics:     [][]byte{[]byte("TOKEN-eeee"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), mectDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"6161616162626262-TOKEN-eeee-02"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsMECTRollback/account-after-create.json"), string(genericResponse.Docs[0].Source))

	// DO ROLLBACK
	err = esProc.RemoveAccountsMECT(5040)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)
}

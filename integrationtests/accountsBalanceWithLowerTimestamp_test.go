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

func TestIndexAccountsBalance(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}

	// ################ UPDATE ACCOUNT-MECT BALANCE ##########################
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: 0,
	}

	body := &dataBlock.Body{}

	mectToken := &mect.MECToken{
		Value: big.NewInt(1000),
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
	accounts = &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
	}

	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("eeeebbbb"),
							Identifier: []byte(core.BuiltInFunctionMECTTransfer),
							Topics:     [][]byte{[]byte("TTTT-abcd"), nil, big.NewInt(1).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"6161616162626262"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-first-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-00"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-mect-first-update.json"), string(genericResponse.Docs[0].Source))

	//////////////////// INDEX BALANCE LOWER TIMESTAMP ///////////////////////////////////

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 5000,
	}
	mockAccount.GetBalanceCalled = func() *big.Int {
		return big.NewInt(1000)
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"6161616162626262"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-first-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-00"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-mect-first-update.json"), string(genericResponse.Docs[0].Source))

	//////////////////// INDEX BALANCE HIGHER TIMESTAMP ///////////////////////////////////
	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 6000,
	}
	mockAccount.GetBalanceCalled = func() *big.Int {
		return big.NewInt(2000)
	}

	pool = &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			"h1": &transaction.Transaction{
				SndAddr: []byte("eeeebbbb"),
			},
		},
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("eeeebbbb"),
							Identifier: []byte(core.BuiltInFunctionMECTTransfer),
							Topics:     [][]byte{[]byte("TTTT-abcd"), nil, big.NewInt(1).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}
	body = &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				Type:     dataBlock.TxBlock,
				TxHashes: [][]byte{[]byte("h1")},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"6161616162626262"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-second-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-00"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-mect-second-update.json"), string(genericResponse.Docs[0].Source))

	//////////////////////// DELETE MECT BALANCE LOWER TIMESTAMP ////////////////

	mectToken.Value = big.NewInt(0)
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

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 6001,
	}
	mockAccount.GetBalanceCalled = func() *big.Int {
		return big.NewInt(2000)
	}

	pool.Txs = make(map[string]coreData.TransactionHandler)
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"6161616162626262"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-second-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-00"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)
}

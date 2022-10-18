//go:build integrationtests

package integrationtests

import (
	"bytes"
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

func TestAccountBalanceNFTTransfer(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}

	// ################ CREATE NFT ##########################
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: 0,
	}

	body := &dataBlock.Body{}

	mectToken := &mect.MECToken{
		Value: big.NewInt(1000),
	}

	addr := "test-address-balance-1"
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
							Address:    []byte("test-address-balance-1"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreate),
							Topics:     [][]byte{[]byte("NFT-abcdef"), big.NewInt(7440483).Bytes(), big.NewInt(1).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"746573742d616464726573732d62616c616e63652d31-NFT-abcdef-718863"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceNftTransfer/balance-nft-after-create.json"), string(genericResponse.Docs[0].Source))

	// ################ TRANSFER NFT ##########################

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
	}

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("test-address-balance-1"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTTransfer),
							Topics:     [][]byte{[]byte("NFT-abcdef"), big.NewInt(7440483).Bytes(), big.NewInt(1).Bytes(), []byte("new-address")},
						},
						nil,
					},
				},
			},
		},
	}

	mectToken = &mect.MECToken{
		Value: big.NewInt(1000),
	}
	addrReceiver := "new-address"
	mockAccountReceiver := &mock.UserAccountStub{
		RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
			return json.Marshal(mectToken)
		},
		AddressBytesCalled: func() []byte {
			return []byte(addrReceiver)
		},
	}
	mockAccountSender := &mock.UserAccountStub{
		RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
			return []byte("{}"), nil
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}
	accounts = &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			if bytes.Equal(container, []byte(addr)) {
				return mockAccountSender, nil
			}
			return mockAccountReceiver, nil
		},
	}
	esProc, err = CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"746573742d616464726573732d62616c616e63652d31-NFT-abcdef-718863"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)

	ids = []string{"6e65772d61646472657373-NFT-abcdef-718863"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsMECTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceNftTransfer/balance-nft-after-transfer.json"), string(genericResponse.Docs[0].Source))
}

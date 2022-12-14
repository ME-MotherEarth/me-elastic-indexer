//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	coreData "github.com/ME-MotherEarth/me-core/data"
	dataBlock "github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/data/indexer"
	"github.com/ME-MotherEarth/me-core/data/smartContractResult"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	indexerdata "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

const (
	expectedMECTTransferTX = `{ "initialPaidFee":"104000110000000","miniBlockHash":"1ecea6dff9ab9a785a2d55720e88c1bbd7d9c56310a035d16163e373879cd0e1","nonce":6,"round":50,"value":"0","receiver":"657264313375377a79656b7a7664767a656b38373638723567617539703636373775667070736a756b6c7539653674377978377268673473363865327a65","sender":"65726431656636343730746a64746c67706139663667336165346e7365646d6a6730677636773733763332787476686b6666663939336871373530786c39","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":104011,"gasUsed":104011,"fee":"104000110000000","data":"RVNEVFRyYW5zZmVyQDU0NDc0ZTJkMzgzODYyMzgzMzY2QDBh","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true,"tokens":["TGN-88b83f"],"mectValues":["10"],"operation":"MECTTransfer"}`
)

func TestMECTTransferTooMuchGasProvided(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("mectTransfer")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	scrHash2 := []byte("scrHash2MECTTransfer")
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash2},
			},
		},
	}

	txMECT := &transaction.Transaction{
		Nonce:    6,
		SndAddr:  []byte("moa1ef6470tjdtlgpa9f6g3ae4nsedmjg0gv6w73v32xtvhkfff993hq750xl9"),
		RcvAddr:  []byte("moa13u7zyekzvdvzek8768r5gau9p6677ufppsjuklu9e6t7yx7rhg4s68e2ze"),
		GasLimit: 104011,
		GasPrice: 1000000000,
		Data:     []byte("MECTTransfer@54474e2d383862383366@0a"),
		Value:    big.NewInt(0),
	}

	scrHash1 := []byte("scrHash1MECTTransfer")
	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          7,
		GasPrice:       1000000000,
		SndAddr:        []byte("moa13u7zyekzvdvzek8768r5gau9p6677ufppsjuklu9e6t7yx7rhg4s68e2ze"),
		RcvAddr:        []byte("moa1ef6470tjdtlgpa9f6g3ae4nsedmjg0gv6w73v32xtvhkfff993hq750xl9"),
		Data:           []byte("@6f6b"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		ReturnMessage:  []byte("@too much gas provided: gas needed = 372000, gas remained = 2250001"),
	}

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          7,
		GasPrice:       1000000000,
		SndAddr:        []byte("moa13u7zyekzvdvzek8768r5gau9p6677ufppsjuklu9e6t7yx7rhg4s68e2ze"),
		RcvAddr:        []byte("moa1ef6470tjdtlgpa9f6g3ae4nsedmjg0gv6w73v32xtvhkfff993hq750xl9"),
		Data:           []byte("@6f6b"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): txMECT,
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash2): scr2,
			string(scrHash1): scr1,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedMECTTransferTX), genericResponse.Docs[0].Source)
}

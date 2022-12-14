//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ME-MotherEarth/me-core/core"
	coreData "github.com/ME-MotherEarth/me-core/data"
	dataBlock "github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/data/indexer"
	"github.com/ME-MotherEarth/me-core/data/smartContractResult"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	indexerdata "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestTransactionWithClaimRewardsGasRefund(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("claimRewards")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	scrHash1 := []byte("scrRefundGasReward")
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: core.MetachainShardId,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}

	refundValue, _ := big.NewInt(0).SetString("49320000000000", 10)
	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          618,
		GasPrice:       1000000000,
		SndAddr:        []byte("moa1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqq8hlllls7a6h81"),
		RcvAddr:        []byte("moa13tfnxanefpjltv9kesf6e6f4n4muvkdqrk0we52nelsjw3lf2t5q8l45u1"),
		Data:           []byte("@6f6b"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		Value:          refundValue,
	}

	rewards, _ := big.NewInt(0).SetString("2932360285576807", 10)
	scrHash2 := []byte("scrRewards")
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		SndAddr:        []byte("moa1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqq8hlllls7a6h81"),
		RcvAddr:        []byte("moa13tfnxanefpjltv9kesf6e6f4n4muvkdqrk0we52nelsjw3lf2t5q8l45u1"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		Value:          rewards,
	}

	tx1 := &transaction.Transaction{
		Nonce:    617,
		SndAddr:  []byte("moa13tfnxanefpjltv9kesf6e6f4n4muvkdqrk0we52nelsjw3lf2t5q8l45u1"),
		RcvAddr:  []byte("moa1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqq8hlllls7a6h81"),
		GasLimit: 6000000,
		GasPrice: 1000000000,
		Data:     []byte("claimRewards"),
		Value:    big.NewInt(0),
	}

	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): tx1,
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash2): scr2,
			string(scrHash1): scr1,
		},
		Logs: []*coreData.LogData{
			{
				TxHash: string(txHash),
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte("writeLog"),
							Topics:     [][]byte{[]byte("something")},
						},
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/claimRewards/tx-claim-rewards.json"),
		string(genericResponse.Docs[0].Source),
	)
}

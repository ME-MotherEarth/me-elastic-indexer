package transactions

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ME-MotherEarth/me-core/core"
	coreData "github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/data/rewardTx"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func createCommonProcessor() dbTransactionBuilder {
	return dbTransactionBuilder{
		addressPubkeyConverter: mock.NewPubkeyConverterMock(32),
		txFeeCalculator: &mock.EconomicsHandlerStub{
			ComputeTxFeeBasedOnGasUsedCalled: func(tx coreData.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
				return big.NewInt(100)
			},
			ComputeGasLimitCalled: func(tx coreData.TransactionWithFeeHandler) uint64 {
				return 500
			},
		},
		shardCoordinator: &mock.ShardCoordinatorMock{},
		dataFieldParser:  createDataFieldParserMock(),
	}
}

func TestGetMoveBalanceTransaction(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}
	status := "Success"
	gasPrice := uint64(1000)
	gasLimit := uint64(1000)
	cp := createCommonProcessor()

	tx := &transaction.Transaction{
		Nonce:       1,
		Value:       big.NewInt(1000),
		RcvAddr:     []byte("receiver"),
		SndAddr:     []byte("sender"),
		GasPrice:    gasPrice,
		GasLimit:    gasLimit,
		Data:        []byte("data"),
		ChainID:     []byte("1"),
		Version:     1,
		Signature:   []byte("signature"),
		RcvUserName: []byte("rcv"),
		SndUserName: []byte("snd"),
	}

	expectedTx := &data.Transaction{
		Hash:                 hex.EncodeToString(txHash),
		MBHash:               hex.EncodeToString(mbHash),
		Nonce:                tx.Nonce,
		Round:                header.Round,
		Value:                tx.Value.String(),
		Receiver:             cp.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:               cp.addressPubkeyConverter.Encode(tx.SndAddr),
		ReceiverShard:        mb.ReceiverShardID,
		SenderShard:          mb.SenderShardID,
		GasPrice:             gasPrice,
		GasLimit:             gasLimit,
		GasUsed:              uint64(500),
		InitialPaidFee:       "100",
		Data:                 tx.Data,
		Signature:            hex.EncodeToString(tx.Signature),
		Timestamp:            time.Duration(header.GetTimeStamp()),
		Status:               status,
		ReceiverAddressBytes: []byte("receiver"),
		Fee:                  "100",
		ReceiverUserName:     []byte("rcv"),
		SenderUserName:       []byte("snd"),
		Operation:            "transfer",
		Version:              1,
		Receivers:            []string{},
	}

	dbTx := cp.prepareTransaction(tx, txHash, mbHash, mb, header, status)
	require.Equal(t, expectedTx, dbTx)
}

func TestGetTransactionByType_RewardTx(t *testing.T) {
	t.Parallel()

	cp := createCommonProcessor()

	round := uint64(10)
	rcvAddr := []byte("receiver")
	rwdTx := &rewardTx.RewardTx{Round: round, RcvAddr: rcvAddr}
	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}
	status := "Success"

	resultTx := cp.prepareRewardTransaction(rwdTx, txHash, mbHash, mb, header, status)
	expectedTx := &data.Transaction{
		Hash:      hex.EncodeToString(txHash),
		MBHash:    hex.EncodeToString(mbHash),
		Round:     round,
		Receiver:  hex.EncodeToString(rcvAddr),
		Status:    status,
		Value:     "<nil>",
		Sender:    fmt.Sprintf("%d", core.MetachainShardId),
		Data:      make([]byte, 0),
		Operation: rewardsOperation,
	}

	require.Equal(t, expectedTx, resultTx)
}

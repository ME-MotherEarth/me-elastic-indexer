package transactions

import (
	"testing"

	coreData "github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/data/receipt"
	"github.com/ME-MotherEarth/me-core/data/rewardTx"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestGroupNormalTxs(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, &mock.EconomicsHandlerStub{}, parser)
	grouper := newTxsGrouper(txBuilder, false, 0, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{txHash1, txHash2},
		Type:     block.TxBlock,
	}
	header := &block.Header{}
	txs := map[string]coreData.TransactionHandler{
		string(txHash1): &transaction.Transaction{
			SndAddr: []byte("sender1"),
			RcvAddr: []byte("receiver1"),
		},
		string(txHash2): &transaction.Transaction{
			SndAddr: []byte("sender2"),
			RcvAddr: []byte("receiver2"),
		},
	}
	alteredAddresses := data.NewAlteredAccounts()

	normalTxs, _ := grouper.groupNormalTxs(0, mb, header, txs, alteredAddresses)
	require.Len(t, normalTxs, 2)
	require.Equal(t, 4, alteredAddresses.Len())
}

func TestGroupRewardsTxs(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, &mock.EconomicsHandlerStub{}, parser)
	grouper := newTxsGrouper(txBuilder, false, 0, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{txHash1, txHash2},
		Type:     block.RewardsBlock,
	}
	header := &block.Header{}
	txs := map[string]coreData.TransactionHandler{
		string(txHash1): &rewardTx.RewardTx{
			RcvAddr: []byte("receiver1"),
		},
		string(txHash2): &rewardTx.RewardTx{
			RcvAddr: []byte("receiver2"),
		},
	}
	alteredAddresses := data.NewAlteredAccounts()

	normalTxs, _ := grouper.groupRewardsTxs(0, mb, header, txs, alteredAddresses)
	require.Len(t, normalTxs, 2)
	require.Equal(t, 2, alteredAddresses.Len())
}

func TestGroupInvalidTxs(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	txBuilder := newTransactionDBBuilder(mock.NewPubkeyConverterMock(32), &mock.ShardCoordinatorMock{}, &mock.EconomicsHandlerStub{}, parser)
	grouper := newTxsGrouper(txBuilder, false, 0, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{txHash1, txHash2},
		Type:     block.InvalidBlock,
	}
	header := &block.Header{}
	txs := map[string]coreData.TransactionHandler{
		string(txHash1): &transaction.Transaction{
			SndAddr: []byte("sender1"),
			RcvAddr: []byte("receiver1"),
		},
		string(txHash2): &transaction.Transaction{
			SndAddr: []byte("sender2"),
			RcvAddr: []byte("receiver2"),
		},
	}
	alteredAddresses := data.NewAlteredAccounts()

	normalTxs, _ := grouper.groupInvalidTxs(0, mb, header, txs, alteredAddresses)
	require.Len(t, normalTxs, 2)
	require.Equal(t, 2, alteredAddresses.Len())
}

func TestGroupReceipts(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, &mock.EconomicsHandlerStub{}, parser)
	grouper := newTxsGrouper(txBuilder, false, 0, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	header := &block.Header{}
	txs := map[string]coreData.TransactionHandler{
		string(txHash1): &receipt.Receipt{
			SndAddr: []byte("sender1"),
		},
		string(txHash2): &receipt.Receipt{
			SndAddr: []byte("sender2"),
		},
	}

	normalTxs := grouper.groupReceipts(header, txs)
	require.Len(t, normalTxs, 2)
}

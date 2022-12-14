package operations

import (
	"testing"

	"github.com/ME-MotherEarth/me-core/data/transaction"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestNewOperationsProcessor(t *testing.T) {
	t.Parallel()

	op, err := NewOperationsProcessor(false, nil)
	require.Nil(t, op)
	require.Equal(t, indexer.ErrNilShardCoordinator, err)

	op, err = NewOperationsProcessor(false, &mock.ShardCoordinatorMock{})
	require.NotNil(t, op)
	require.Nil(t, err)
}

func TestOperationsProcessor_ProcessTransactionsAndSCRSTransactions(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor(true, &mock.ShardCoordinatorMock{})

	txs := []*data.Transaction{
		{},
		{
			ReceiverShard: 1,
		},
	}

	processedTxs, _ := op.ProcessTransactionsAndSCRs(txs, nil)
	require.Equal(t, []*data.Transaction{
		{Type: string(transaction.TxTypeNormal)},
	}, processedTxs)
}

func TestOperationsProcessor_ProcessTransactionsAndSCRSSmartContractResults(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor(true, &mock.ShardCoordinatorMock{})

	scrs := []*data.ScResult{
		{},
		{
			ReceiverShard: 1,
		},
	}

	_, processedSCRs := op.ProcessTransactionsAndSCRs(nil, scrs)
	require.Equal(t, []*data.ScResult{
		{Type: string(transaction.TxTypeUnsigned), Status: transaction.TxStatusSuccess.String()},
	}, processedSCRs)
}

func TestOperationsProcessor_ShouldIgnoreSCRs(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor(true, &mock.ShardCoordinatorMock{})

	scrs := []*data.ScResult{
		{
			ReturnMessage: data.GasRefundForRelayerMessage,
			Data:          nil,
		},
		{
			Data: []byte("@6f6b"),
		},
		{
			Operation:          "MECTNFTTransfer",
			SenderAddressBytes: []byte("sender"),
		},
	}

	_, processedSCRs := op.ProcessTransactionsAndSCRs(nil, scrs)
	for _, scr := range processedSCRs {
		require.True(t, scr.CanBeIgnored)
	}
}

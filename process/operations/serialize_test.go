package operations

import (
	"testing"

	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestOperationsProcessor_SerializeSCRS(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor(false, &mock.ShardCoordinatorMock{})

	scrs := []*data.ScResult{
		{
			SenderShard:   0,
			ReceiverShard: 1,
		},
		{
			SenderShard:   2,
			ReceiverShard: 0,
		},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := op.SerializeSCRs(scrs, buffSlice, "operations")
	require.Nil(t, err)
	require.Equal(t, `{"update":{"_index":"operations","_id":""}}
{"script":{"source":"return"},"upsert":{"nonce":0,"gasLimit":0,"gasPrice":0,"value":"","sender":"","receiver":"","senderShard":0,"receiverShard":1,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}}
{ "index" : { "_index":"operations","_id" : "" } }
{"nonce":0,"gasLimit":0,"gasPrice":0,"value":"","sender":"","receiver":"","senderShard":2,"receiverShard":0,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
`, buffSlice.Buffers()[0].String())
}

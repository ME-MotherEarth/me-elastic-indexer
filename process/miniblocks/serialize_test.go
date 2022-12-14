package miniblocks

import (
	"testing"

	"github.com/ME-MotherEarth/me-core/data/block"

	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestMiniblocksProcessor_SerializeBulkMiniBlocks(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(0, mock.HasherMock{}, &mock.MarshalizerMock{}, false)

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1},
		{Hash: "h2", SenderShardID: 0, ReceiverShardID: 2},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, nil, buffSlice, "miniblocks")

	expectedBuff := `{ "index" : { "_index":"miniblocks", "_id" : "h1"} }
{"senderShard":0,"receiverShard":1,"type":"","timestamp":0}
{ "index" : { "_index":"miniblocks", "_id" : "h2"} }
{"senderShard":0,"receiverShard":2,"type":"","timestamp":0}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestMiniblocksProcessor_SerializeBulkMiniBlocksInDB(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(0, mock.HasherMock{}, &mock.MarshalizerMock{}, false)

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1},
		{Hash: "h2", SenderShardID: 0, ReceiverShardID: 2},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, map[string]bool{
		"h1": true,
	}, buffSlice, "miniblocks")

	expectedBuff := `{ "update" : {"_index":"miniblocks", "_id" : "h1" } }
{ "doc" : { "senderBlockHash" : "", "procTypeS": "" } }
{ "index" : { "_index":"miniblocks", "_id" : "h2"} }
{"senderShard":0,"receiverShard":2,"type":"","timestamp":0}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestSerializeMiniblock_CrossShardNormal(t *testing.T) {
	mp, _ := NewMiniblocksProcessor(1, mock.HasherMock{}, &mock.MarshalizerMock{}, false)

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 0, ReceiverShardID: 1, ReceiverBlockHash: "receiverBlock"},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, map[string]bool{
		"h1": true,
	}, buffSlice, "miniblocks")

	expectedBuff := `{ "update" : {"_index":"miniblocks", "_id" : "h1" } }
{ "doc" : { "receiverBlockHash" : "receiverBlock", "procTypeD": "" } }
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

func TestSerializeMiniblock_IntraShardScheduled(t *testing.T) {
	mp, _ := NewMiniblocksProcessor(1, mock.HasherMock{}, &mock.MarshalizerMock{}, false)

	miniblocks := []*data.Miniblock{
		{Hash: "h1", SenderShardID: 1, ReceiverShardID: 1, SenderBlockHash: "senderBlock",
			ProcessingTypeOnSource: block.Scheduled.String()},
	}

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, map[string]bool{
		"h1": false,
	}, buffSlice, "miniblocks")

	expectedBuff := `{ "index" : { "_index":"miniblocks", "_id" : "h1"} }
{"senderShard":1,"receiverShard":1,"senderBlockHash":"senderBlock","type":"","procTypeS":"Scheduled","timestamp":0}
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())

	miniblocks = []*data.Miniblock{
		{Hash: "h1", SenderShardID: 1, ReceiverShardID: 1, ReceiverBlockHash: "receiverBlock",
			ProcessingTypeOnDestination: block.Processed.String()},
	}

	buffSlice = data.NewBufferSlice(data.DefaultMaxBulkSize)
	mp.SerializeBulkMiniBlocks(miniblocks, map[string]bool{
		"h1": true,
	}, buffSlice, "miniblocks")

	expectedBuff = `{ "update" : {"_index":"miniblocks", "_id" : "h1" } }
{ "doc" : { "receiverBlockHash" : "receiverBlock", "procTypeD": "Processed" } }
`
	require.Equal(t, expectedBuff, buffSlice.Buffers()[0].String())
}

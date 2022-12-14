package miniblocks

import (
	"testing"

	dataBlock "github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/hashing"
	"github.com/ME-MotherEarth/me-core/marshal"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestNewMiniblocksProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  func() (uint32, hashing.Hasher, marshal.Marshalizer, bool)
		exErr error
	}{
		{
			name: "NilHash",
			args: func() (uint32, hashing.Hasher, marshal.Marshalizer, bool) {
				return 0, nil, &mock.MarshalizerMock{}, false
			},
			exErr: indexer.ErrNilHasher,
		},
		{
			name: "NilMarshalizer",
			args: func() (uint32, hashing.Hasher, marshal.Marshalizer, bool) {
				return 0, &mock.HasherMock{}, nil, false
			},
			exErr: indexer.ErrNilMarshalizer,
		},
	}

	for _, test := range tests {
		_, err := NewMiniblocksProcessor(test.args())
		require.Equal(t, test.exErr, err)
	}
}

func TestMiniblocksProcessor_PrepareDBMiniblocks(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(0, &mock.HasherMock{}, &mock.MarshalizerMock{}, false)

	header := &dataBlock.Header{}
	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 0,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 0,
			},
		},
	}

	miniblocks := mp.PrepareDBMiniblocks(header, body)
	require.Len(t, miniblocks, 3)
}

func TestMiniblocksProcessor_PrepareScheduledMB(t *testing.T) {
	t.Parallel()

	marshalizer := &marshal.GogoProtoMarshalizer{}
	mp, _ := NewMiniblocksProcessor(0, &mock.HasherMock{}, marshalizer, false)

	mbhr := &dataBlock.MiniBlockHeaderReserved{
		ExecutionType: dataBlock.ProcessingType(1),
	}

	mbhrBytes, _ := marshalizer.Marshal(mbhr)

	header := &dataBlock.Header{
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{
				Reserved: []byte{0},
			},
			{
				Reserved: mbhrBytes,
			},
		},
	}
	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
		},
	}

	miniblocks := mp.PrepareDBMiniblocks(header, body)
	require.Len(t, miniblocks, 2)
	require.Equal(t, dataBlock.Scheduled.String(), miniblocks[1].ProcessingTypeOnSource)
}

func TestMiniblocksProcessor_GetMiniblocksHashesHexEncoded(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(0, &mock.HasherMock{}, &mock.MarshalizerMock{}, false)

	header := &dataBlock.Header{
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{}, {}, {},
		},
	}
	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 2,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 0,
			},
		},
	}

	expectedHashes := []string{
		"c57392e53257b4861f5e406349a8deb89c6dbc2127564ee891a41a188edbf01a",
		"28fda294dc987e5099d75e53cd6f87a9a42b96d55242a634385b5d41175c0c21",
		"44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
	}
	miniblocksHashes := mp.GetMiniblocksHashesHexEncoded(header, body)
	require.Equal(t, expectedHashes, miniblocksHashes)
}

func TestMiniblocksProcessor_GetMiniblocksHashesHexEncodedImportDBMode(t *testing.T) {
	t.Parallel()

	mp, _ := NewMiniblocksProcessor(1, &mock.HasherMock{}, &mock.MarshalizerMock{}, true)

	header := &dataBlock.Header{
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{}, {}, {},
		},
		ShardID: 1,
	}
	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				SenderShardID:   1,
				ReceiverShardID: 2,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 0,
			},
			{
				SenderShardID:   1,
				ReceiverShardID: 1,
			},
		},
	}

	expectedHashes := []string{
		"4a270e1ddac6b429c14c7ebccdcdd53e4f68aeebfc41552c775a7f5a5c35d06d",
	}
	miniblocksHashes := mp.GetMiniblocksHashesHexEncoded(header, body)
	require.Equal(t, expectedHashes, miniblocksHashes)
}

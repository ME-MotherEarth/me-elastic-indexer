package logsevents

import (
	"math/big"
	"testing"

	"github.com/ME-MotherEarth/me-core/data/transaction"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestProcessNFTProperties_Update(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte("MECTNFTUpdateAttributes"),
		Topics:     [][]byte{[]byte("TOUC-aaaa"), big.NewInt(1).Bytes(), nil, []byte("new-something")},
	}
	args := &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	nftsPropertiesP := newNFTsPropertiesProcessor(&mock.PubkeyConverterMock{})

	res := nftsPropertiesP.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.NFTDataUpdate{
		Identifier:    "TOUC-aaaa-01",
		NewAttributes: []byte("new-something"),
		Address:       "61646472",
	}, res.updatePropNFT)
}

func TestProcessNFTProperties_AddUris(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte("MECTNFTAddURI"),
		Topics:     [][]byte{[]byte("TOUC-aaaa"), big.NewInt(1).Bytes(), nil, []byte("uri1"), []byte("uri2")},
	}
	args := &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	nftsPropertiesP := newNFTsPropertiesProcessor(&mock.PubkeyConverterMock{})

	res := nftsPropertiesP.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.NFTDataUpdate{
		Identifier: "TOUC-aaaa-01",
		URIsToAdd:  [][]byte{[]byte("uri1"), []byte("uri2")},
		Address:    "61646472",
	}, res.updatePropNFT)
}

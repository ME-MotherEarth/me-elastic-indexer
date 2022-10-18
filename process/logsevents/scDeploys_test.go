package logsevents

import (
	"testing"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestScDeploysProcessor(t *testing.T) {
	t.Parallel()

	scDeploysProc := newSCDeploysProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.SCDeployIdentifier),
		Topics:     [][]byte{[]byte("addr1"), []byte("addr2")},
	}

	scDeploys := map[string]*data.ScDeployInfo{}
	res := scDeploysProc.processEvent(&argsProcessEvent{
		event:            event,
		timestamp:        1000,
		scDeploys:        scDeploys,
		txHashHexEncoded: "01020304",
	})
	require.True(t, res.processed)

	require.Equal(t, &data.ScDeployInfo{
		TxHash:    "01020304",
		Creator:   "6164647232",
		Timestamp: uint64(1000),
	}, scDeploys["6164647231"])
}

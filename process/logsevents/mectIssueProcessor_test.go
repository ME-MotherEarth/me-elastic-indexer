package logsevents

import (
	"testing"
	"time"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestIssueMECTProcessor(t *testing.T) {
	t.Parallel()

	mectIssueProc := newMECTIssueProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(issueNonFungibleMECTFunc),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), []byte("my-token"), []byte("MYTOKEN"), []byte(core.NonFungibleMECT)},
	}
	args := &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	res := mectIssueProc.processEvent(args)

	require.Equal(t, &data.TokenInfo{
		Token:        "MYTOKEN-abcd",
		Name:         "my-token",
		Ticker:       "MYTOKEN",
		Timestamp:    time.Duration(1234),
		Type:         core.NonFungibleMECT,
		Issuer:       "61646472",
		CurrentOwner: "61646472",
		OwnersHistory: []*data.OwnerData{
			{
				Address:   "61646472",
				Timestamp: time.Duration(1234),
			},
		},
		Properties: &data.TokenProperties{},
	}, res.tokenInfo)
}

func TestIssueMECTProcessor_TransferOwnership(t *testing.T) {
	t.Parallel()

	mectIssueProc := newMECTIssueProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(transferOwnershipFunc),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), []byte("my-token"), []byte("MYTOKEN"), []byte(core.NonFungibleMECT), []byte("newOwner")},
	}
	args := &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	res := mectIssueProc.processEvent(args)

	require.Equal(t, &data.TokenInfo{
		Token:        "MYTOKEN-abcd",
		Name:         "my-token",
		Ticker:       "MYTOKEN",
		Timestamp:    time.Duration(1234),
		Type:         core.NonFungibleMECT,
		Issuer:       "61646472",
		CurrentOwner: "6e65774f776e6572",
		OwnersHistory: []*data.OwnerData{
			{
				Address:   "6e65774f776e6572",
				Timestamp: time.Duration(1234),
			},
		},
		TransferOwnership: true,
		Properties:        &data.TokenProperties{},
	}, res.tokenInfo)
}

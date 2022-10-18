package workItems_test

import (
	"errors"
	"testing"

	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/ME-MotherEarth/me-elastic-indexer/workItems"
	"github.com/stretchr/testify/require"
)

func TestItemRounds_Save(t *testing.T) {
	called := false
	itemRounds := workItems.NewItemRounds(
		&mock.ElasticProcessorStub{
			SaveRoundsInfoCalled: func(infos []*data.RoundInfo) error {
				called = true
				return nil
			},
		},
		[]*data.RoundInfo{
			{},
		},
	)
	require.False(t, itemRounds.IsInterfaceNil())

	err := itemRounds.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemRounds_SaveRoundsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemRounds := workItems.NewItemRounds(
		&mock.ElasticProcessorStub{
			SaveRoundsInfoCalled: func(infos []*data.RoundInfo) error {
				return localErr
			},
		},
		[]*data.RoundInfo{
			{},
		},
	)
	require.False(t, itemRounds.IsInterfaceNil())

	err := itemRounds.Save()
	require.Equal(t, localErr, err)
}

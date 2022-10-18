package workItems_test

import (
	"errors"
	"testing"

	coreData "github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/ME-MotherEarth/me-elastic-indexer/workItems"
	"github.com/stretchr/testify/require"
)

func TestItemAccounts_Save(t *testing.T) {
	called := false
	itemAccounts := workItems.NewItemAccounts(
		&mock.ElasticProcessorStub{
			SaveAccountsCalled: func(_ uint64, _ []*data.Account) error {
				called = true
				return nil
			},
		},
		0,
		[]coreData.UserAccountHandler{},
	)
	require.False(t, itemAccounts.IsInterfaceNil())

	err := itemAccounts.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemAccounts_SaveAccountsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemAccounts := workItems.NewItemAccounts(
		&mock.ElasticProcessorStub{
			SaveAccountsCalled: func(_ uint64, _ []*data.Account) error {
				return localErr
			},
		},
		0,
		[]coreData.UserAccountHandler{},
	)
	require.False(t, itemAccounts.IsInterfaceNil())

	err := itemAccounts.Save()
	require.Equal(t, localErr, err)
}

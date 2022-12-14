package workItems_test

import (
	"errors"
	"testing"

	"github.com/ME-MotherEarth/me-core/data"
	dataBlock "github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/ME-MotherEarth/me-elastic-indexer/workItems"
	"github.com/stretchr/testify/require"
)

func TestItemRemoveBlock_Save(t *testing.T) {
	countCalled := 0
	itemRemove := workItems.NewItemRemoveBlock(
		&mock.ElasticProcessorStub{
			RemoveHeaderCalled: func(header data.HeaderHandler) error {
				countCalled++
				return nil
			},
			RemoveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				countCalled++
				return nil
			},
			RemoveTransactionsCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				countCalled++
				return nil
			},
		},
		&dataBlock.Body{},
		&dataBlock.Header{},
	)
	require.False(t, itemRemove.IsInterfaceNil())

	err := itemRemove.Save()
	require.NoError(t, err)
	require.Equal(t, 3, countCalled)
}

func TestItemRemoveBlock_SaveRemoveHeaderShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemRemove := workItems.NewItemRemoveBlock(
		&mock.ElasticProcessorStub{
			RemoveHeaderCalled: func(header data.HeaderHandler) error {
				return localErr
			},
		},
		&dataBlock.Body{},
		&dataBlock.Header{},
	)
	require.False(t, itemRemove.IsInterfaceNil())

	err := itemRemove.Save()
	require.Equal(t, localErr, err)
}

func TestItemRemoveBlock_SaveRemoveMiniblocksShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemRemove := workItems.NewItemRemoveBlock(
		&mock.ElasticProcessorStub{
			RemoveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				return localErr
			},
		},
		&dataBlock.Body{},
		&dataBlock.Header{},
	)
	require.False(t, itemRemove.IsInterfaceNil())

	err := itemRemove.Save()
	require.Equal(t, localErr, err)
}

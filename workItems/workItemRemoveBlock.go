package workItems

import (
	"github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-core/data/block"
)

type itemRemoveBlock struct {
	indexer       removeIndexer
	bodyHandler   data.BodyHandler
	headerHandler data.HeaderHandler
}

// NewItemRemoveBlock will create a new instance of itemRemoveBlock
func NewItemRemoveBlock(
	indexer removeIndexer,
	bodyHandler data.BodyHandler,
	headerHandler data.HeaderHandler,
) WorkItemHandler {
	return &itemRemoveBlock{
		indexer:       indexer,
		bodyHandler:   bodyHandler,
		headerHandler: headerHandler,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (wirb *itemRemoveBlock) IsInterfaceNil() bool {
	return wirb == nil
}

// Save will remove a block and miniblocks from elasticsearch database
func (wirb *itemRemoveBlock) Save() error {
	err := wirb.indexer.RemoveHeader(wirb.headerHandler)
	if err != nil {
		return err
	}

	body, ok := wirb.bodyHandler.(*block.Body)
	if !ok {
		return ErrBodyTypeAssertion
	}

	err = wirb.indexer.RemoveMiniblocks(wirb.headerHandler, body)
	if err != nil {
		return err
	}

	err = wirb.indexer.RemoveTransactions(wirb.headerHandler, body)
	if err != nil {
		return err
	}

	return wirb.indexer.RemoveAccountsMECT(wirb.headerHandler.GetTimeStamp())
}

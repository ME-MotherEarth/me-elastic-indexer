package modifiers

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/transactions"
)

type responseSCRsBulk struct {
	Hits struct {
		Hits []struct {
			ID     string         `json:"_id"`
			Source *data.ScResult `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type scrsModifier struct {
	pubKeyConverter     core.PubkeyConverter
	operationDataParser transactions.DataFieldParser
}

// NewSCRsModifier will create a new instance of scrsModifier
func NewSCRsModifier() (*scrsModifier, error) {
	pubKeyConverter, parser, err := createPubKeyConverterAndParser()
	if err != nil {
		return nil, err
	}

	return &scrsModifier{
		pubKeyConverter:     pubKeyConverter,
		operationDataParser: parser,
	}, nil
}

// Modify will modify the smart contract results from the provided responseBody
func (sm *scrsModifier) Modify(responseBody []byte) ([]*bytes.Buffer, error) {
	responseSCRs := &responseSCRsBulk{}
	err := json.Unmarshal(responseBody, responseSCRs)
	if err != nil {
		return nil, err
	}

	buffSlice := data.NewBufferSlice()
	for _, hit := range responseSCRs.Hits.Hits {
		if shouldIgnoreSCR(hit.Source) {
			continue
		}

		errPrep := sm.prepareSCRForIndexing(hit.Source)
		if errPrep != nil {
			log.Warn("cannot prepare scr",
				"error", errPrep.Error(),
				"hash", hit.ID,
			)
			continue
		}

		meta, serializedData, errSerialize := serializeSCR(hit.ID, hit.Source)
		if errSerialize != nil {
			log.Warn("cannot serialize scr",
				"error", errSerialize.Error(),
				"hash", hit.ID,
			)
			continue
		}

		errPut := buffSlice.PutData(meta, serializedData)
		if errPut != nil {
			log.Warn("cannot put scr",
				"error", errPut.Error(),
				"hash", hit.ID,
			)
			continue
		}
	}

	return buffSlice.Buffers(), nil
}

func shouldIgnoreSCR(scr *data.ScResult) bool {
	if scr.Status == "pending" {
		return true
	}

	return false
}

func serializeSCR(hash string, scr *data.ScResult) ([]byte, []byte, error) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, hash, "\n"))
	serializedData, errPrepareReceipt := json.Marshal(scr)
	if errPrepareReceipt != nil {
		return nil, nil, errPrepareReceipt
	}

	return meta, serializedData, nil
}

func (sm *scrsModifier) prepareSCRForIndexing(scr *data.ScResult) error {
	sndAddr, err := sm.pubKeyConverter.Decode(scr.Sender)
	if err != nil {
		return err
	}
	rcvAddr, err := sm.pubKeyConverter.Decode(scr.Receiver)
	if err != nil {
		return err
	}

	res := sm.operationDataParser.Parse(scr.Data, sndAddr, rcvAddr)

	// TODO uncomment this when create index `operations`
	//scr.Type = string(transaction.TxTypeUnsigned)
	//scr.Status = transaction.TxStatusSuccess.String()

	scr.Operation = res.Operation
	scr.Function = res.Function
	scr.MECTValues = res.MECTValues
	scr.Tokens = res.Tokens
	scr.Receivers = res.Receivers
	scr.ReceiversShardIDs = res.ReceiversShardID

	return nil
}

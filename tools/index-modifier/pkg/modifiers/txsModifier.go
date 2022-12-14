package modifiers

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/pubkeyConverter"
	factoryMarshalizer "github.com/ME-MotherEarth/me-core/marshal/factory"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/transactions"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/transactions/datafield"
	"github.com/ME-MotherEarth/me-elastic-indexer/tools/index-modifier/pkg/modifiers/utils"
	logger "github.com/ME-MotherEarth/me-logger"
)

var log = logger.GetOrCreate("index-modifier/pkg/alterindex")

type responseTransactionsBulk struct {
	Hits struct {
		Hits []struct {
			ID     string            `json:"_id"`
			Source *data.Transaction `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type txsModifier struct {
	pubKeyConverter     core.PubkeyConverter
	operationDataParser transactions.DataFieldParser
}

// NewTxsModifier will create a new instance of txsModifier
func NewTxsModifier() (*txsModifier, error) {
	pubKeyConverter, parser, err := createPubKeyConverterAndParser()
	if err != nil {
		return nil, err
	}

	return &txsModifier{
		pubKeyConverter:     pubKeyConverter,
		operationDataParser: parser,
	}, nil
}

func createOperationParser(pubkeyConverter core.PubkeyConverter) (transactions.DataFieldParser, error) {
	shardCoordinator, err := utils.NewMultiShardCoordinator(3, 0)
	if err != nil {
		return nil, err
	}
	marshalizer, err := factoryMarshalizer.NewMarshalizer(factoryMarshalizer.GogoProtobuf)

	arguments := &datafield.ArgsOperationDataFieldParser{
		PubKeyConverter:  pubkeyConverter,
		Marshalizer:      marshalizer,
		ShardCoordinator: shardCoordinator,
	}

	return datafield.NewOperationDataFieldParser(arguments)
}

func createPubKeyConverterAndParser() (core.PubkeyConverter, transactions.DataFieldParser, error) {
	pubKeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	if err != nil {
		return nil, nil, err
	}

	parser, err := createOperationParser(pubKeyConverter)
	if err != nil {
		return nil, nil, err
	}

	return pubKeyConverter, parser, nil
}

// Modify will modify the transactions from the provided responseBody
func (tm *txsModifier) Modify(responseBody []byte) ([]*bytes.Buffer, error) {
	responseTxs := &responseTransactionsBulk{}
	err := json.Unmarshal(responseBody, responseTxs)
	if err != nil {
		return nil, err
	}

	buffSlice := data.NewBufferSlice()
	for _, hit := range responseTxs.Hits.Hits {
		if shouldIgnoreTx(hit.Source) {
			continue
		}

		errPrep := tm.prepareTxForIndexing(hit.Source)
		if errPrep != nil {
			log.Warn("cannot prepare transaction",
				"error", errPrep.Error(),
				"hash", hit.ID,
			)
			continue
		}

		meta, serializedData, errSerialize := serializeTx(hit.ID, hit.Source)
		if errSerialize != nil {
			log.Warn("cannot serialize transaction",
				"error", errSerialize.Error(),
				"hash", hit.ID,
			)
			continue
		}

		errPut := buffSlice.PutData(meta, serializedData)
		if errPut != nil {
			log.Warn("cannot put transaction",
				"error", errPut.Error(),
				"hash", hit.ID,
			)
			continue
		}
	}

	return buffSlice.Buffers(), nil
}

func shouldIgnoreTx(tx *data.Transaction) bool {
	if tx.Status == "pending" {
		return true
	}

	return false
}

func serializeTx(hash string, tx *data.Transaction) ([]byte, []byte, error) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, hash, "\n"))
	serializedData, errPrepareReceipt := json.Marshal(tx)
	if errPrepareReceipt != nil {
		return nil, nil, errPrepareReceipt
	}

	return meta, serializedData, nil
}

func (tm *txsModifier) prepareTxForIndexing(tx *data.Transaction) error {
	if tx.Sender == "4294967295" {
		// TODO uncomment this when create index `operations`
		// tx.Type = string(transaction.TxTypeNormal)
		return nil
	}

	sndAddr, err := tm.pubKeyConverter.Decode(tx.Sender)
	if err != nil {
		return err
	}
	rcvAddr, err := tm.pubKeyConverter.Decode(tx.Receiver)
	if err != nil {
		return err
	}

	res := tm.operationDataParser.Parse(tx.Data, sndAddr, rcvAddr)

	// TODO uncomment this when create index `operations`
	// tx.Type = string(transaction.TxTypeNormal)

	tx.Operation = res.Operation
	tx.Function = res.Function
	tx.MECTValues = res.MECTValues
	tx.Tokens = res.Tokens
	tx.Receivers = res.Receivers
	tx.ReceiversShardIDs = res.ReceiversShardID

	return nil
}

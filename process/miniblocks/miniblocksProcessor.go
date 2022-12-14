package miniblocks

import (
	"encoding/hex"
	"time"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/check"
	coreData "github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/hashing"
	"github.com/ME-MotherEarth/me-core/marshal"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	logger "github.com/ME-MotherEarth/me-logger"
)

var log = logger.GetOrCreate("indexer/process/miniblocks")

type miniblocksProcessor struct {
	hasher       hashing.Hasher
	marshalier   marshal.Marshalizer
	selfShardID  uint32
	importDBMode bool
}

// NewMiniblocksProcessor will create a new instance of miniblocksProcessor
func NewMiniblocksProcessor(
	selfShardID uint32,
	hasher hashing.Hasher,
	marshalier marshal.Marshalizer,
	isImportDBMode bool,
) (*miniblocksProcessor, error) {
	if check.IfNil(marshalier) {
		return nil, indexer.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, indexer.ErrNilHasher
	}

	return &miniblocksProcessor{
		hasher:       hasher,
		marshalier:   marshalier,
		selfShardID:  selfShardID,
		importDBMode: isImportDBMode,
	}, nil
}

// PrepareDBMiniblocks will prepare miniblocks from body
func (mp *miniblocksProcessor) PrepareDBMiniblocks(header coreData.HeaderHandler, body *block.Body) []*data.Miniblock {
	headerHash, err := mp.calculateHash(header)
	if err != nil {
		log.Warn("indexer: could not calculate header hash", "error", err)
		return nil
	}

	dbMiniblocks := make([]*data.Miniblock, 0)
	for mbIndex, miniblock := range body.MiniBlocks {
		dbMiniblock, errPrepareMiniblock := mp.prepareMiniblockForDB(mbIndex, miniblock, header, headerHash)
		if errPrepareMiniblock != nil {
			log.Warn("miniblocksProcessor.PrepareDBMiniblocks cannot prepare miniblock", "error", errPrepareMiniblock)
			continue
		}

		dbMiniblocks = append(dbMiniblocks, dbMiniblock)
	}

	return dbMiniblocks
}

func (mp *miniblocksProcessor) prepareMiniblockForDB(
	mbIndex int,
	miniblock *block.MiniBlock,
	header coreData.HeaderHandler,
	headerHash []byte,
) (*data.Miniblock, error) {
	mbHash, err := mp.calculateHash(miniblock)
	if err != nil {
		return nil, err
	}

	encodedMbHash := hex.EncodeToString(mbHash)

	dbMiniblock := &data.Miniblock{
		Hash:            encodedMbHash,
		SenderShardID:   miniblock.SenderShardID,
		ReceiverShardID: miniblock.ReceiverShardID,
		Type:            miniblock.Type.String(),
		Timestamp:       time.Duration(header.GetTimeStamp()),
		Reserved:        miniblock.Reserved,
	}

	encodedHeaderHash := hex.EncodeToString(headerHash)
	isIntraShard := dbMiniblock.SenderShardID == dbMiniblock.ReceiverShardID
	isCrossOnSource := !isIntraShard && dbMiniblock.SenderShardID == header.GetShardID()
	if isIntraShard || isCrossOnSource {
		mp.setFieldsMBIntraShardAndCrossFromMe(mbIndex, header, encodedHeaderHash, dbMiniblock)

		return dbMiniblock, nil
	}

	processingType, _ := mp.computeProcessingTypeAndConstructionState(mbIndex, header)
	dbMiniblock.ProcessingTypeOnDestination = processingType
	dbMiniblock.ReceiverBlockHash = encodedHeaderHash

	return dbMiniblock, nil
}

func (mp *miniblocksProcessor) setFieldsMBIntraShardAndCrossFromMe(
	mbIndex int,
	header coreData.HeaderHandler,
	headerHash string,
	dbMiniblock *data.Miniblock,
) {
	processingType, constructionState := mp.computeProcessingTypeAndConstructionState(mbIndex, header)

	dbMiniblock.ProcessingTypeOnSource = processingType
	switch {
	case constructionState == int32(block.Final) && processingType == block.Normal.String():
		dbMiniblock.SenderBlockHash = headerHash
		dbMiniblock.ReceiverBlockHash = headerHash
		dbMiniblock.ProcessingTypeOnSource = processingType
		dbMiniblock.ProcessingTypeOnDestination = processingType
	case constructionState == int32(block.Proposed) && processingType == block.Scheduled.String():
		dbMiniblock.SenderBlockHash = headerHash
		dbMiniblock.ProcessingTypeOnSource = processingType
	case constructionState == int32(block.Final) && processingType == block.Processed.String():
		dbMiniblock.ReceiverBlockHash = headerHash
		dbMiniblock.ProcessingTypeOnDestination = processingType
	}
}

func (mp *miniblocksProcessor) computeProcessingTypeAndConstructionState(mbIndex int, header coreData.HeaderHandler) (string, int32) {
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniblockHeaders) <= mbIndex {
		return "", 0
	}

	processingType := miniblockHeaders[mbIndex].GetProcessingType()
	constructionState := miniblockHeaders[mbIndex].GetConstructionState()

	switch processingType {
	case int32(block.Scheduled):
		return block.Scheduled.String(), constructionState
	case int32(block.Processed):
		return block.Processed.String(), constructionState
	default:
		return block.Normal.String(), constructionState
	}
}

// GetMiniblocksHashesHexEncoded will compute miniblocks hashes in a hexadecimal encoding
func (mp *miniblocksProcessor) GetMiniblocksHashesHexEncoded(header coreData.HeaderHandler, body *block.Body) []string {
	if body == nil || len(header.GetMiniBlockHeadersHashes()) == 0 {
		return nil
	}

	encodedMiniblocksHashes := make([]string, 0)
	selfShardID := header.GetShardID()
	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type == block.PeerBlock {
			continue
		}

		isCrossShard := miniblock.ReceiverShardID != miniblock.SenderShardID
		shouldIgnore := selfShardID == miniblock.SenderShardID && mp.importDBMode && isCrossShard
		if shouldIgnore {
			continue
		}

		isDstMe := selfShardID == miniblock.ReceiverShardID
		if isDstMe && isCrossShard {
			continue
		}

		miniblockHash, err := mp.calculateHash(miniblock)
		if err != nil {
			log.Debug("miniblocksProcessor.GetMiniblocksHashesHexEncoded cannot calculate miniblock hash",
				"error", err)
			continue
		}
		encodedMiniblocksHashes = append(encodedMiniblocksHashes, hex.EncodeToString(miniblockHash))
	}

	return encodedMiniblocksHashes
}

func (mp *miniblocksProcessor) calculateHash(object interface{}) ([]byte, error) {
	return core.CalculateHash(mp.marshalier, mp.hasher, object)
}

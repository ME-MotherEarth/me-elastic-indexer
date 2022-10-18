package logsevents

import (
	coreData "github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/tokeninfo"
)

type argsProcessEvent struct {
	txHashHexEncoded        string
	scDeploys               map[string]*data.ScDeployInfo
	txs                     map[string]*data.Transaction
	event                   coreData.EventHandler
	accounts                data.AlteredAccountsHandler
	tokens                  data.TokensHandler
	tokensSupply            data.TokensHandler
	tokenRolesAndProperties *tokeninfo.TokenRolesAndProperties
	timestamp               uint64
	logAddress              []byte
}

type argOutputProcessEvent struct {
	identifier      string
	value           string
	receiver        string
	receiverShardID uint32
	tokenInfo       *data.TokenInfo
	delegator       *data.Delegator
	processed       bool
	updatePropNFT   *data.NFTDataUpdate
}

type eventsProcessor interface {
	processEvent(args *argsProcessEvent) argOutputProcessEvent
}

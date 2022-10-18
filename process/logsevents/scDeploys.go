package logsevents

import (
	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
)

type scDeploysProcessor struct {
	scDeploysIdentifiers map[string]struct{}
	pubKeyConverter      core.PubkeyConverter
}

func newSCDeploysProcessor(pubKeyConverter core.PubkeyConverter) *scDeploysProcessor {
	return &scDeploysProcessor{
		pubKeyConverter: pubKeyConverter,
		scDeploysIdentifiers: map[string]struct{}{
			core.SCDeployIdentifier:  {},
			core.SCUpgradeIdentifier: {},
		},
	}
}

func (sdp *scDeploysProcessor) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	eventIdentifier := string(args.event.GetIdentifier())
	_, ok := sdp.scDeploysIdentifiers[eventIdentifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	topics := args.event.GetTopics()
	if len(topics) < 2 {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	scAddress := sdp.pubKeyConverter.Encode(topics[0])
	args.scDeploys[scAddress] = &data.ScDeployInfo{
		TxHash:    args.txHashHexEncoded,
		Creator:   sdp.pubKeyConverter.Encode(topics[1]),
		Timestamp: args.timestamp,
	}

	return argOutputProcessEvent{
		processed: true,
	}
}

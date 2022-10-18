package logsevents

import (
	"unicode"

	"github.com/ME-MotherEarth/me-core/core"
)

const (
	tokenTopicsIndex            = 0
	propertyPairStep            = 2
	mectPropertiesStartIndex    = 2
	minTopicsPropertiesAndRoles = 4
	upgradePropertiesEvent      = "upgradeProperties"
)

type mectPropertiesProc struct {
	pubKeyConverter            core.PubkeyConverter
	rolesOperationsIdentifiers map[string]struct{}
}

func newMectPropertiesProcessor(pubKeyConverter core.PubkeyConverter) *mectPropertiesProc {
	return &mectPropertiesProc{
		pubKeyConverter: pubKeyConverter,
		rolesOperationsIdentifiers: map[string]struct{}{
			core.BuiltInFunctionSetMECTRole:               {},
			core.BuiltInFunctionUnSetMECTRole:             {},
			core.BuiltInFunctionMECTNFTCreateRoleTransfer: {},
			upgradePropertiesEvent:                        {},
		},
	}
}

func (epp *mectPropertiesProc) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	identifier := string(args.event.GetIdentifier())
	_, ok := epp.rolesOperationsIdentifiers[identifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	topics := args.event.GetTopics()
	if len(topics) < minTopicsPropertiesAndRoles {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	if identifier == upgradePropertiesEvent {
		return epp.extractTokenProperties(args)
	}

	if identifier == core.BuiltInFunctionMECTNFTCreateRoleTransfer {
		return epp.extractDataNFTCreateRoleTransfer(args)
	}

	// topics contains:
	// [0] --> token identifier
	// [1] --> nonce of the NFT (bytes)
	// [2] --> value
	// [3:] --> roles to set or unset

	rolesBytes := topics[3:]
	ok = checkRolesBytes(rolesBytes)
	if !ok {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	shouldAddRole := identifier == core.BuiltInFunctionSetMECTRole
	addrBech := epp.pubKeyConverter.Encode(args.event.GetAddress())
	for _, roleBytes := range rolesBytes {
		args.tokenRolesAndProperties.AddRole(string(topics[tokenTopicsIndex]), addrBech, string(roleBytes), shouldAddRole)
	}

	return argOutputProcessEvent{
		processed: true,
	}
}

func (epp *mectPropertiesProc) extractDataNFTCreateRoleTransfer(args *argsProcessEvent) argOutputProcessEvent {
	topics := args.event.GetTopics()

	addrBech := epp.pubKeyConverter.Encode(args.event.GetAddress())
	shouldAddCreateRole := bytesToBool(topics[3])
	args.tokenRolesAndProperties.AddRole(string(topics[tokenTopicsIndex]), addrBech, core.MECTRoleNFTCreate, shouldAddCreateRole)

	return argOutputProcessEvent{
		processed: true,
	}
}

func (epp *mectPropertiesProc) extractTokenProperties(args *argsProcessEvent) argOutputProcessEvent {
	topics := args.event.GetTopics()
	properties := topics[mectPropertiesStartIndex:]
	propertiesMap := make(map[string]bool)
	for i := 0; i < len(properties); i += propertyPairStep {
		property := string(properties[i])
		val := bytesToBool(properties[i+1])
		propertiesMap[property] = val
	}

	args.tokenRolesAndProperties.AddProperties(string(topics[tokenTopicsIndex]), propertiesMap)

	return argOutputProcessEvent{
		processed: true,
	}
}

func checkRolesBytes(rolesBytes [][]byte) bool {
	for _, role := range rolesBytes {
		if !containsNonLetterChars(string(role)) {
			return false
		}
	}

	return true
}

func containsNonLetterChars(data string) bool {
	for _, c := range data {
		if !unicode.IsLetter(c) {
			return false
		}
	}

	return true
}

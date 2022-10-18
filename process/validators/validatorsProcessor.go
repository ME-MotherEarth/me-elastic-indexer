package validators

import (
	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/check"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
)

type validatorsProcessor struct {
	bulkSizeMaxSize          int
	validatorPubkeyConverter core.PubkeyConverter
}

// NewValidatorsProcessor will create a new instance of validatorsProcessor
func NewValidatorsProcessor(validatorPubkeyConverter core.PubkeyConverter, bulkSizeMaxSize int) (*validatorsProcessor, error) {
	if check.IfNil(validatorPubkeyConverter) {
		return nil, indexer.ErrNilPubkeyConverter
	}

	return &validatorsProcessor{
		bulkSizeMaxSize:          bulkSizeMaxSize,
		validatorPubkeyConverter: validatorPubkeyConverter,
	}, nil
}

// PrepareValidatorsPublicKeys will prepare validators public keys
func (vp *validatorsProcessor) PrepareValidatorsPublicKeys(shardValidatorsPubKeys [][]byte) *data.ValidatorsPublicKeys {
	validatorsPubKeys := &data.ValidatorsPublicKeys{
		PublicKeys: make([]string, 0),
	}

	for _, validatorPk := range shardValidatorsPubKeys {
		strValidatorPk := vp.validatorPubkeyConverter.Encode(validatorPk)

		validatorsPubKeys.PublicKeys = append(validatorsPubKeys.PublicKeys, strValidatorPk)
	}

	return validatorsPubKeys
}

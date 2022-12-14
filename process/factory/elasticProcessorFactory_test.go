package factory

import (
	"testing"

	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateElasticProcessor(t *testing.T) {

	args := ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		DBClient:                 &mock.DatabaseWriterStub{},
		AccountsDB:               &mock.AccountsStub{},
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		TransactionFeeCalculator: &mock.EconomicsHandlerStub{},
		EnabledIndexes:           []string{"blocks"},
		Denomination:             1,
		IsInImportDBMode:         false,
		UseKibana:                false,
	}

	ep, err := CreateElasticProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, ep)
}

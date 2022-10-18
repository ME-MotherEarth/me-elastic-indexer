//go:build integrationtests

package integrationtests

import (
	"math/big"
	"testing"

	"github.com/ME-MotherEarth/me-core/core"
	coreData "github.com/ME-MotherEarth/me-core/data"
	dataBlock "github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/data/indexer"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	indexerdata "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestIssueTokenAndSetRole(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("TOK-abcd"), []byte("semi-token"), []byte("SEMI"), []byte(core.SemiFungibleMECT)},
						},
						{
							Address:    []byte("addr"),
							Identifier: []byte("upgradeProperties"),
							Topics:     [][]byte{[]byte("TOK-abcd"), big.NewInt(0).Bytes(), []byte("canUpgrade"), []byte("true")},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"TOK-abcd"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/issueTokenAndSetRoles/token-after-issue-ok.json"), string(genericResponse.Docs[0].Source))

	// SET ROLES
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("my-address"),
							Identifier: []byte(core.BuiltInFunctionSetMECTRole),
							Topics:     [][]byte{[]byte("TOK-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.MECTRoleNFTCreate), []byte(core.MECTRoleNFTBurn)},
						},
						nil,
					},
				},
			},
		},
	}

	header.TimeStamp = 10000
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"TOK-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/issueTokenAndSetRoles/token-after-set-role.json"), string(genericResponse.Docs[0].Source))

	// TRANSFER ROLE
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("my-address"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreateRoleTransfer),
							Topics:     [][]byte{[]byte("TOK-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte("false")},
						},
						{
							Address:    []byte("new-address"),
							Identifier: []byte(core.BuiltInFunctionMECTNFTCreateRoleTransfer),
							Topics:     [][]byte{[]byte("TOK-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte("true")},
						},
					},
				},
			},
		},
	}

	header.TimeStamp = 10000
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"TOK-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/issueTokenAndSetRoles/token-after-transfer-role.json"), string(genericResponse.Docs[0].Source))

	// UNSET ROLES
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("my-address"),
							Identifier: []byte(core.BuiltInFunctionUnSetMECTRole),
							Topics:     [][]byte{[]byte("TOK-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.MECTRoleNFTBurn)},
						},
						nil,
					},
				},
			},
		},
	}

	header.TimeStamp = 10000
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"TOK-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/issueTokenAndSetRoles/token-after-unset-role.json"), string(genericResponse.Docs[0].Source))
}

func TestIssueSetRolesEventAndAfterTokenIssue(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}

	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	// SET ROLES
	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("my-address"),
							Identifier: []byte(core.BuiltInFunctionSetMECTRole),
							Topics:     [][]byte{[]byte("TTT-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.MECTRoleNFTCreate), []byte(core.MECTRoleNFTBurn)},
						},
						nil,
					},
				},
			},
		},
	}

	header.TimeStamp = 10000
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"TTT-abcd"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/issueTokenAndSetRoles/token-after-set-roles-first.json"), string(genericResponse.Docs[0].Source))

	// ISSUE
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("TTT-abcd"), []byte("semi-token"), []byte("SEMI"), []byte(core.SemiFungibleMECT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"TTT-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/issueTokenAndSetRoles/token-after-set-roles-and-issue.json"), string(genericResponse.Docs[0].Source))
}

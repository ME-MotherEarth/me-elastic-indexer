package logsevents

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/data/transaction"
	"github.com/ME-MotherEarth/me-elastic-indexer/mock"
	"github.com/ME-MotherEarth/me-elastic-indexer/process/tokeninfo"
	"github.com/stretchr/testify/require"
)

func TestMectPropertiesProcCreateRoleShouldWork(t *testing.T) {
	t.Parallel()

	mectPropProc := newMectPropertiesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionSetMECTRole),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.MECTRoleNFTCreate)},
	}

	tokenRolesAndProperties := tokeninfo.NewTokenRolesAndProperties()
	mectPropProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := map[string][]*tokeninfo.RoleData{
		core.MECTRoleNFTCreate: {
			{
				Token:   "MYTOKEN-abcd",
				Set:     true,
				Address: "61646472",
			},
		},
	}
	require.Equal(t, expected, tokenRolesAndProperties.GetRoles())
}

func TestMectPropertiesProcTransferCreateRole(t *testing.T) {
	t.Parallel()

	mectPropProc := newMectPropertiesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionMECTNFTCreateRoleTransfer),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(strconv.FormatBool(true))},
	}

	tokenRolesAndProperties := tokeninfo.NewTokenRolesAndProperties()
	mectPropProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := map[string][]*tokeninfo.RoleData{
		core.MECTRoleNFTCreate: {
			{
				Token:   "MYTOKEN-abcd",
				Set:     true,
				Address: "61646472",
			},
		},
	}
	require.Equal(t, expected, tokenRolesAndProperties.GetRoles())
}

func TestMectPropertiesProcUpgradeProperties(t *testing.T) {
	t.Parallel()

	mectPropProc := newMectPropertiesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(upgradePropertiesEvent),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), []byte("canMint"), []byte("true"), []byte("canBurn"), []byte("false")},
	}

	tokenRolesAndProperties := tokeninfo.NewTokenRolesAndProperties()
	mectPropProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := []*tokeninfo.PropertiesData{
		{
			Token: "MYTOKEN-abcd",
			Properties: map[string]bool{
				"canMint": true,
				"canBurn": false,
			},
		},
	}
	require.Equal(t, expected, tokenRolesAndProperties.GetAllTokensWithProperties())
}

func TestCheckRolesBytes(t *testing.T) {
	t.Parallel()

	role1, _ := hex.DecodeString("01")
	role2, _ := hex.DecodeString("02")
	rolesBytes := [][]byte{role1, role2}
	require.False(t, checkRolesBytes(rolesBytes))

	role1 = []byte("MECTRoleNFTCreate")
	rolesBytes = [][]byte{role1}
	require.True(t, checkRolesBytes(rolesBytes))
}

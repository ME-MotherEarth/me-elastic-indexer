package noKibana

import (
	"github.com/ME-MotherEarth/me-elastic-indexer/templates"
)

type Array = templates.Array
type Object = templates.Object

// Accounts will hold the configuration for the accounts index
var Accounts = Object{
	"index_patterns": Array{
		"accounts-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
	},

	"mappings": Object{
		"properties": Object{
			"balanceNum": Object{
				"type": "double",
			},
			"totalBalanceWithStakeNum": Object{
				"type": "double",
			},
			"nonce": Object{
				"type": "double",
			},
			"address": Object{
				"type": "text",
			},
			"balance": Object{
				"type": "text",
			},
			"totalBalanceWithStake": Object{
				"type": "text",
			},
		},
	},
}

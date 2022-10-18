package withKibana

var AccountsMECTHistory = Object{
	"index_patterns": Array{
		"accountsmecthistory-*",
	},
	"settings": Object{
		"number_of_shards":   5,
		"number_of_replicas": 0,
		"opendistro.index_state_management.rollover_alias": "accountsmecthistory",
	},
	"mappings": Object{
		"properties": Object{
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
			"tokenNonce": Object{
				"type": "double",
			},
		},
	},
}

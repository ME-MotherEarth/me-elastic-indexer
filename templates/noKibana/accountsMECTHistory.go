package noKibana

// AccountsMECTHistory will hold the configuration for the accountsmecthistory index
var AccountsMECTHistory = Object{
	"index_patterns": Array{
		"accountsmecthistory-*",
	},
	"settings": Object{
		"number_of_shards":   5,
		"number_of_replicas": 0,
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

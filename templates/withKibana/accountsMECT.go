package withKibana

// AccountsMECT will hold the configuration for the accountsmect index
var AccountsMECT = Object{
	"index_patterns": Array{
		"accountsmect-*",
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
			"data": Object{
				"type": "nested",
				"properties": Object{
					"name": Object{
						"type": "text",
					},
					"creator": Object{
						"type": "text",
					},
					"tags": Object{
						"type": "text",
					},
					"attributes": Object{
						"type": "text",
					},
					"metadata": Object{
						"type": "text",
					},
				},
			},
			"tokenNonce": Object{
				"type": "double",
			},
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
		},
	},
}

package templatesAndPolicies

import (
	"bytes"

	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/templates/noKibana"
)

type templatesAndPolicyReaderNoKibana struct{}

// NewTemplatesAndPolicyReaderNoKibana will create a new instance of templatesAndPolicyReaderNoKibana
func NewTemplatesAndPolicyReaderNoKibana() *templatesAndPolicyReaderNoKibana {
	return new(templatesAndPolicyReaderNoKibana)
}

// GetElasticTemplatesAndPolicies will return templates and policies
func (tr *templatesAndPolicyReaderNoKibana) GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	indexPolicies := make(map[string]*bytes.Buffer)
	indexTemplates := make(map[string]*bytes.Buffer)

	indexTemplates["opendistro"] = noKibana.OpenDistro.ToBuffer()
	indexTemplates[indexer.TransactionsIndex] = noKibana.Transactions.ToBuffer()
	indexTemplates[indexer.BlockIndex] = noKibana.Blocks.ToBuffer()
	indexTemplates[indexer.MiniblocksIndex] = noKibana.Miniblocks.ToBuffer()
	indexTemplates[indexer.RatingIndex] = noKibana.Rating.ToBuffer()
	indexTemplates[indexer.RoundsIndex] = noKibana.Rounds.ToBuffer()
	indexTemplates[indexer.ValidatorsIndex] = noKibana.Validators.ToBuffer()
	indexTemplates[indexer.AccountsIndex] = noKibana.Accounts.ToBuffer()
	indexTemplates[indexer.AccountsHistoryIndex] = noKibana.AccountsHistory.ToBuffer()
	indexTemplates[indexer.AccountsMECTIndex] = noKibana.AccountsMECT.ToBuffer()
	indexTemplates[indexer.AccountsMECTHistoryIndex] = noKibana.AccountsMECTHistory.ToBuffer()
	indexTemplates[indexer.EpochInfoIndex] = noKibana.EpochInfo.ToBuffer()
	indexTemplates[indexer.ReceiptsIndex] = noKibana.Receipts.ToBuffer()
	indexTemplates[indexer.ScResultsIndex] = noKibana.SCResults.ToBuffer()
	indexTemplates[indexer.SCDeploysIndex] = noKibana.SCDeploys.ToBuffer()
	indexTemplates[indexer.TokensIndex] = noKibana.Tokens.ToBuffer()
	indexTemplates[indexer.TagsIndex] = noKibana.Tags.ToBuffer()
	indexTemplates[indexer.LogsIndex] = noKibana.Logs.ToBuffer()
	indexTemplates[indexer.DelegatorsIndex] = noKibana.Delegators.ToBuffer()
	indexTemplates[indexer.OperationsIndex] = noKibana.Operations.ToBuffer()
	indexTemplates[indexer.CollectionsIndex] = noKibana.Collections.ToBuffer()

	return indexTemplates, indexPolicies, nil
}

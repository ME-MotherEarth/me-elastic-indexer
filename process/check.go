package process

import (
	"github.com/ME-MotherEarth/me-core/core/check"
	elasticIndexer "github.com/ME-MotherEarth/me-elastic-indexer"
)

func checkArguments(arguments *ArgElasticProcessor) error {
	if arguments == nil {
		return elasticIndexer.ErrNilElasticProcessorArguments
	}
	if arguments.EnabledIndexes == nil {
		return elasticIndexer.ErrNilEnabledIndexesMap
	}
	if check.IfNilReflect(arguments.DBClient) {
		return elasticIndexer.ErrNilDatabaseClient
	}
	if check.IfNilReflect(arguments.StatisticsProc) {
		return elasticIndexer.ErrNilStatisticHandler
	}
	if check.IfNilReflect(arguments.BlockProc) {
		return elasticIndexer.ErrNilBlockHandler
	}
	if check.IfNilReflect(arguments.AccountsProc) {
		return elasticIndexer.ErrNilAccountsHandler
	}
	if check.IfNilReflect(arguments.MiniblocksProc) {
		return elasticIndexer.ErrNilMiniblocksHandler
	}
	if check.IfNilReflect(arguments.ValidatorsProc) {
		return elasticIndexer.ErrNilValidatorsHandler
	}
	if arguments.TransactionsProc == nil {
		return elasticIndexer.ErrNilTransactionsHandler
	}
	if check.IfNilReflect(arguments.LogsAndEventsProc) {
		return elasticIndexer.ErrNilLogsAndEventsHandler
	}
	if check.IfNilReflect(arguments.OperationsProc) {
		return elasticIndexer.ErrNilOperationsHandler
	}

	return nil
}

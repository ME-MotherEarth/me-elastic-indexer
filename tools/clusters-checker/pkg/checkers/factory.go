package checkers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ME-MotherEarth/me-elastic-indexer/tools/clusters-checker/pkg/client"
	"github.com/ME-MotherEarth/me-elastic-indexer/tools/clusters-checker/pkg/config"
	"github.com/elastic/go-elasticsearch/v7"
)

// CreateClusterChecker will create a new instance of clusterChecker structure
func CreateClusterChecker(cfg *config.Config, timestampIndex int, logPrefix string) (*clusterChecker, error) {
	clientSource, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{cfg.SourceCluster.URL},
		Username:  cfg.SourceCluster.User,
		Password:  cfg.SourceCluster.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create source client %s", err.Error())
	}

	clientDestination, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{cfg.DestinationCluster.URL},
		Username:  cfg.DestinationCluster.User,
		Password:  cfg.DestinationCluster.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create destination client %s", err.Error())
	}

	return &clusterChecker{
		clientSource:         clientSource,
		clientDestination:    clientDestination,
		indicesWithTimestamp: cfg.Compare.IndicesWithTimestamp,
		indicesNoTimestamp:   cfg.Compare.IndicesNoTimestamp,

		missingFromSource:      map[string]json.RawMessage{},
		missingFromDestination: map[string]json.RawMessage{},

		startTimestamp: cfg.Compare.IntervalSettings[timestampIndex].Start,
		stopTimestamp:  cfg.Compare.IntervalSettings[timestampIndex].Stop,
		logPrefix:      logPrefix,
	}, nil
}

func CreateMultipleCheckers(cfg *config.Config) ([]*clusterChecker, error) {
	checkers := make([]*clusterChecker, 0, len(cfg.Compare.IntervalSettings))

	for idx := 0; idx < len(cfg.Compare.IntervalSettings); idx++ {
		logPrefix := "instance_" + strconv.FormatUint(uint64(idx), 10)
		cc, err := CreateClusterChecker(cfg, idx, logPrefix)
		if err != nil {
			return nil, err
		}

		checkers = append(checkers, cc)
	}

	return checkers, nil
}

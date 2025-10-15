package factory

import (
	"fmt"

	"github.com/edsonmichaque/bazaruto/pkg/job"
	"github.com/edsonmichaque/bazaruto/pkg/job/adapter"
)

// CreateAdapter creates the appropriate job adapter based on configuration
func CreateAdapter(adapterType job.AdapterType, config interface{}) (job.Adapter, error) {
	switch adapterType {
	case job.AdapterTypeMemory:
		return adapter.NewMemoryAdapter(), nil
	case job.AdapterTypeRedis:
		redisConfig, ok := config.(adapter.RedisAdapterConfig)
		if !ok {
			return nil, fmt.Errorf("invalid Redis adapter config")
		}
		return adapter.NewRedisAdapter(redisConfig)
	case job.AdapterTypeDatabase:
		databaseConfig, ok := config.(adapter.DatabaseAdapterConfig)
		if !ok {
			return nil, fmt.Errorf("invalid database adapter config")
		}
		return adapter.NewDatabaseAdapter(databaseConfig)
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", adapterType)
	}
}

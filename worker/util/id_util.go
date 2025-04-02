package util

import (
	idgenerator "synapse/common/id-generator"
	"synapse/worker/config"
)

func GetNextId() int64 {
	return idgenerator.MustGenerateSnowflakeIDWithInstanceKey(config.SnowflakeNodeIDRedisKeyForNextId)
}

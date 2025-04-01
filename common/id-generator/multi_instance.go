package id_generator

import (
	"context"
	"synapse/common/log"

	"github.com/bwmarrin/snowflake"
	"github.com/redis/go-redis/v9"
)

var (
	instancesMap = make(map[string]*snowflake.Node)
)

func InitMultiSnowflakeInstances(ctx context.Context, redisCli redis.UniversalClient, keys ...string) error {
	snowflake.Epoch = 1724728815000
	for _, k := range keys {
		nextNodeID, err := redisCli.Incr(ctx, k).Uint64()
		if err != nil {
			return err
		}
		nextNodeID -= 1 // 0 can be used
		nextNodeID %= 1024
		log.Log.Infof("Snowflake node id initialized(key=%v): %v\n", k, nextNodeID)
		flakeNode, err := snowflake.NewNode(int64(nextNodeID))
		if err != nil {
			return err
		}
		instancesMap[k] = flakeNode
	}
	return nil
}

func MustGenerateSnowflakeIDWithInstanceKey(instanceKey string) int64 {
	n, ok := instancesMap[instanceKey]
	if !ok {
		panic("MustGenerateSnowflakeIDWithInstanceKey: key not exists: " + instanceKey)
	}
	return n.Generate().Int64()
}

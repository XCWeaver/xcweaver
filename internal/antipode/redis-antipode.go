package antipode

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func CreateRedis(redis_host string, redis_port string, redis_password string) Redis {
	return Redis{redis.NewClient(&redis.Options{
		Addr:     redis_host + ":" + redis_port,
		Password: redis_password,
		DB:       0, // use default DB
	})}
}

func (r Redis) write(ctx context.Context, _ string, key string, obj AntiObj) error {

	jsonAntiObj, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	err = r.client.Set(ctx, key, jsonAntiObj, 0).Err()

	return err
}

func (r Redis) read(ctx context.Context, _ string, key string) (AntiObj, error) {

	jsonAntiObj, err := r.client.Get(ctx, key).Bytes()

	if err != nil {
		return AntiObj{}, err
	}

	var obj AntiObj
	err = json.Unmarshal(jsonAntiObj, &obj)
	if err == redis.Nil {
		return AntiObj{}, ErrNotFound
	} else if err != nil {
		return AntiObj{}, err
	}

	return obj, err
}

func (r Redis) consume(context.Context, string, string, chan struct{}) (<-chan AntiObj, error) {
	return nil, nil
}

func (r Redis) barrier(ctx context.Context, lineage []WriteIdentifier, datastoreID string) error {

	for _, writeIdentifier := range lineage {
		if writeIdentifier.Dtstid == datastoreID {
			for {
				jsonAntiObj, err := r.client.Get(ctx, writeIdentifier.Key).Bytes()

				if !errors.Is(err, redis.Nil) && err != nil {
					return err
				} else if errors.Is(err, redis.Nil) { //the version replication process is not yet completed
					continue
				} else {
					var obj AntiObj
					err = json.Unmarshal(jsonAntiObj, &obj)
					if err != nil {
						return err
					} else if obj.Version == writeIdentifier.Version { //the version replication process is already completed
						break
					} else { //the version replication process is not yet completed
						continue
					}
				}
			}
		}
	}

	return nil
}

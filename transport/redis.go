package transport

import (
	"context"
	"flag"
	"strconv"

	flowmessage "github.com/cloudflare/goflow/v3/pb"
	"github.com/cloudflare/goflow/v3/utils"
	"github.com/redis/go-redis/v9"
)

var (
	RedisAddress *string
	RedisPasswd  *string
	RedisDB      *int
)

type RedisClient struct {
	Client     *redis.Client
	log        utils.Logger
	input      chan *flowmessage.FlowMessage
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func RegisterFlags() {
	RedisAddress = flag.String("redis.address", "localhost:6379", "Address of the redis server")
	RedisPasswd = flag.String("redis.passwd", "", "Password for redis authentication")
	RedisDB = flag.Int("redis.db", 0, "Db index of the redis database")
}
func StartRedisClientFromArgs(log utils.Logger) (*RedisClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := redis.NewClient(&redis.Options{
		Addr:     *RedisAddress,
		Password: *RedisPasswd,
		DB:       *RedisDB,
	})
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	log.Infof("Connected to Redis: %s", pong)
	ch := make(chan *flowmessage.FlowMessage)
	/*go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case pair := <-ch:
				err := client.Set(ctx, pair.Key, pair.Value, 0).Err()
				if err != nil {
					fmt.Println("Error setting key:", err)
				}
				fmt.Printf("Persisted key: %s, value: %s\n", pair.Key, pair.Value)
			}
		}
	}()*/
	return &RedisClient{Client: client, log: log, ctx: ctx, cancelFunc: cancel, input: ch}, nil
}
func (rc RedisClient) Publish(msgs []*flowmessage.FlowMessage) {
	for _, msg := range msgs {
		rc.Client.Set(rc.ctx, "flow:"+strconv.FormatUint(uint64(msg.SequenceNum), 10), msg, 90)
	}
}

func (rc *RedisClient) Close() error {
	return rc.Client.Close()
}

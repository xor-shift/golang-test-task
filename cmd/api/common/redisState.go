package common

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/streadway/amqp"
	"log"
	"os"
	"sync"
)

type RedisState struct {
	waitGroup   *sync.WaitGroup
	RedisClient *redis.Client
}

func NewRedisState() *RedisState {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return &RedisState{
		waitGroup:   &sync.WaitGroup{},
		RedisClient: redisClient,
	}
}

func (state *RedisState) Close() error {
	return state.RedisClient.Close()
}

func (state *RedisState) Wait() {
	state.waitGroup.Wait()
}

func (state *RedisState) RunOne(channel <-chan amqp.Delivery) {
	state.waitGroup.Add(1)
	go func() {
		defer state.waitGroup.Done()

		for delivery := range channel {
			log.Printf("New delivery with content-type of %s", delivery.ContentType)

			body := delivery.Body
			message, err := MessageFromJSON(body)
			if err != nil {
				log.Printf("Failed to unmarshal a queue message with the body: %s", string(body))
			}

			key, value := message.RedisMarshal()
			log.Printf("%s, %s", key, value)
			if err := state.RedisClient.LPush(key, value).Err(); err != nil {
				_ = fmt.Errorf("failed to push the kv pair of (%s, %s) to redis", key, value)
				os.Exit(1)
			}
		}
	}()
}

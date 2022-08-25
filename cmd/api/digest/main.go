package main

import (
	"fmt"
	dotenv "github.com/joho/godotenv"
	"os"
	"strconv"
	"twitch_chat_analysis/cmd/api/common"
)

func main() {
	if err := dotenv.Load(); err != nil {
		_ = fmt.Errorf("failed to dotenv.Load with errror: %v", err)
		os.Exit(1)
	}

	//state, err := common.NewMQState("localhost", 7001, "user", "password")
	mqPort, _ := strconv.ParseUint(os.Getenv("MQ_PORT"), 10, 16)
	state, err := common.NewMQState(
		os.Getenv("MQ_HOST"),
		uint16(mqPort),
		os.Getenv("MQ_USER"),
		os.Getenv("MQ_PASSWORD"),
	)
	if err != nil {
		_ = fmt.Errorf("failed to create an IngestState with error: %v\n", err)
		os.Exit(1)
	}
	defer state.EndState()

	//redisState := common.NewRedisState("localhost", 6379, "", 0)
	redisPort, _ := strconv.ParseUint(os.Getenv("REDIS_PORT"), 10, 16)
	redisDB, _ := strconv.ParseInt(os.Getenv("REDIS_DB"), 10, 32)
	redisState := common.NewRedisState(os.Getenv("REDIS_HOST"), uint16(redisPort), os.Getenv("REDIS_USER"), int(redisDB))
	defer func() {
		if err := redisState.Close(); err != nil {
			_ = fmt.Errorf("error while closing redis client: %v", err)
		}
	}()

	consumeChannel, err := state.Consume()
	if err != nil {
		_ = fmt.Errorf("failed to create a consumer channel to the database\n")
	}

	redisState.RunOne(consumeChannel)
	redisState.Wait()
}

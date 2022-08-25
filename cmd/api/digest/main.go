package main

import (
	"fmt"
	"os"
	"twitch_chat_analysis/cmd/api/common"
)

func main() {
	state, err := common.NewMQState()
	if err != nil {
		_ = fmt.Errorf("failed to create an IngestState with error: %v\n", err)
		os.Exit(1)
	}
	defer state.EndState()

	redisState := common.NewRedisState()
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

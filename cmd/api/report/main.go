package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"twitch_chat_analysis/cmd/api/common"
)

type Request struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
}

func main() {
	r := gin.Default()

	redisState := common.NewRedisState()
	defer func() {
		if err := redisState.Close(); err != nil {
			_ = fmt.Errorf("error while closing redis client: %v", err)
		}
	}()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, "worked")
	})

	r.GET("/message/list", func(c *gin.Context) {
		request := Request{}
		if err := c.BindJSON(&request); err != nil {
			if err = c.AbortWithError(http.StatusBadRequest, errors.New("malformed or invalid request")); err != nil {
				_ = fmt.Errorf("error while calling AbortWithError: %v", err)
			}
			return
		}

		key := common.RedisMarshalSenderReceiver(request.Sender, request.Receiver)
		result, err := redisState.RedisClient.LRange(key, 0, -1).Result()
		if err != nil {
			_ = fmt.Errorf("internal error while getting LRange: %v", err)
			if err = c.AbortWithError(http.StatusInternalServerError, errors.New("internal error")); err != nil {
				_ = fmt.Errorf("error while calling AbortWithError: %v", err)
			}
			return
		}
		c.JSON(http.StatusOK, map[string]any{"messages": result})
	})

	if err := r.Run(":8081"); err != nil {
		panic(err)
	}
}

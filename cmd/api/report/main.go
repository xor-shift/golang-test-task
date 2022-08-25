package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	dotenv "github.com/joho/godotenv"
	"net/http"
	"os"
	"strconv"
	"twitch_chat_analysis/cmd/api/common"
)

type Request struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
}

func main() {
	if err := dotenv.Load(); err != nil {
		_ = fmt.Errorf("failed to dotenv.Load with errror: %v", err)
		os.Exit(1)
	}

	r := gin.Default()

	//redisState := common.NewRedisState("localhost", 6379, "", 0)
	redisPort, _ := strconv.ParseUint(os.Getenv("REDIS_PORT"), 10, 16)
	redisDB, _ := strconv.ParseInt(os.Getenv("REDIS_DB"), 10, 32)
	redisState := common.NewRedisState(os.Getenv("REDIS_HOST"), uint16(redisPort), os.Getenv("REDIS_USER"), int(redisDB))
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

	if err := r.Run(fmt.Sprintf(":%s", os.Getenv("REPORT_SERVE_PORT"))); err != nil {
		panic(err)
	}
}

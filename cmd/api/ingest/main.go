package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"twitch_chat_analysis/cmd/api/common"
)

func main() {
	r := gin.Default()

	state, err := common.NewMQState()
	if err != nil {
		_ = fmt.Errorf("failed to create an IngestState with error: %v", err)
		os.Exit(1)
	}

	defer state.EndState()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, "worked")
	})

	r.POST("/message", func(c *gin.Context) {
		var err error

		message := common.Message{}

		if err = c.BindJSON(&message); err != nil {
			if err = c.AbortWithError(http.StatusBadRequest, errors.New("malformed or invalid JSON")); err != nil {
				_ = fmt.Errorf("error while calling AbortWithError: %v", err)
			}
			return
		}

		if err = state.NewMessage(message); err != nil {
			_ = fmt.Errorf("internal error while pushing new message: %v", err)
			if err = c.AbortWithError(http.StatusInternalServerError, errors.New("internal error")); err != nil {
				_ = fmt.Errorf("error while calling AbortWithError: %v", err)
			}
			return
		}
	})

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

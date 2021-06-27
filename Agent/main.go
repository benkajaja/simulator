package main

import (
	"math/rand"
	"simulator/Agent/conf"
	"simulator/Agent/objdetectmod"
	"simulator/Agent/status"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	r := gin.Default()
	r.GET("status", status.Statuscheck)
	objDetectModService := r.Group("/objdetectmod")
	objDetectModService.GET("/init", objdetectmod.Init)
	objDetectModService.POST("/inference", objdetectmod.Inference)
	objDetectModService.POST("/upload", objdetectmod.Upload)
	r.Run("0.0.0.0:" + conf.AGENT_PORT)
}

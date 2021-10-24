package main

import (
	"log"
	"math/rand"
	"simulator/Agent/conf"
	"simulator/Agent/objdetectmod"
	"simulator/Agent/status"
	"simulator/Agent/visualnavigationmod"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := conf.Init("./conf.json"); err != nil {
		log.Fatal("[ERROR] load conf.json fail", err)
	}
	r := gin.Default()
	r.GET("probe", status.Probe)
	objDetectModService := r.Group("/objdetectmod")
	objDetectModService.GET("/init", objdetectmod.Init)
	objDetectModService.POST("/inference", objdetectmod.Inference)
	objDetectModService.POST("/upload", objdetectmod.Upload)
	objDetectModService.GET("/policy", objdetectmod.PolicyGET)
	objDetectModService.POST("/policy", objdetectmod.PolicyPOST)
	objDetectModService.GET("/tasknum", objdetectmod.TaskNumGET)

	visualNavigationModService := r.Group("/visualnavigationmod")
	visualNavigationModService.GET("/init", visualnavigationmod.Init)
	visualNavigationModService.POST("/inference", visualnavigationmod.Inference)
	visualNavigationModService.POST("/upload", visualnavigationmod.Upload)
	visualNavigationModService.GET("/policy", visualnavigationmod.PolicyGET)
	visualNavigationModService.POST("/policy", visualnavigationmod.PolicyPOST)
	visualNavigationModService.GET("/tasknum", visualnavigationmod.TaskNumGET)
	r.Run("0.0.0.0:" + conf.AGENT_PORT)
}

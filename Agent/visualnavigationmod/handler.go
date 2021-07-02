package visualnavigationmod

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"simulator/Agent/conf"
	service "simulator/Agent/visualnavigationgrpc"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func Init(g *gin.Context) {
	addr := "0.0.0.0:" + conf.VISUAL_NAVIGATION_MOD_SERVICE_PORT
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Println("[ERROR] Can not connect to gRPC server: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}
	defer conn.Close()

	c := service.NewVisualNavigationModserviceClient(conn)
	r, err := c.Init(context.Background(),
		&service.InitArg{
			Input:          "input/Cut2-030-U3_01_20190817_012500.mp4",
			Sec:            0.9,
			ObjDetectModel: "model/frozen_inference_graph.pb",
			SiameseModel:   "model/Siamese_tracking.ckpt",
		},
	)
	if err != nil {
		log.Println("[ERROR] Can not connect to gRPC server: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}
	log.Println("[DEBUG] init result: ", r.Status)
	g.JSON(http.StatusOK, gin.H{"message": r.Status})
}

func Inference(g *gin.Context) {
	dst := "./upload/"
	file, err := g.FormFile("file")
	if err != nil {
		log.Println("[ERROR] FormFile err: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}
	// file.Filename abc.mp4
	err = g.SaveUploadedFile(file, dst+file.Filename)
	if err != nil {
		log.Println("[ERROR] SaveUploadedFile err: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}
	ext := filepath.Ext(file.Filename)            // .mp4
	key := strings.TrimSuffix(file.Filename, ext) // abc
	targetName := key + "-" + conf.ROLE + ext
	sourceVideoPath, _ := filepath.Abs(dst + file.Filename)
	outputVideoPath, _ := filepath.Abs(dst + targetName)
	outputDirPath, _ := filepath.Abs(dst)

	var score float32
	var action string
	if conf.VISUAL_NAVIGATION_MOD_POLICY == "SAVE" || conf.ROLE == "CLOUD" {
		score, action, err = localInference(outputDirPath, sourceVideoPath, outputVideoPath)
		if err != nil {
			g.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	} else {
		action = "VISUAL CLOUD INFERENCE"
		score, err = cloudInference(sourceVideoPath)
		if err != nil {
			g.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	g.JSON(http.StatusOK, gin.H{
		"message": "OK",
		"score":   score,
		"action":  action,
	})
}

func Upload(g *gin.Context) {
	dst := "./upload/"
	file, err := g.FormFile("file")
	if err != nil {
		log.Println("[ERROR] FormFile err: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	// file.Filename abc.mp4
	ext := filepath.Ext(file.Filename)            // .mp4
	key := strings.TrimSuffix(file.Filename, ext) // abc
	targetName := key + "-visualnavigation" + ext
	err = g.SaveUploadedFile(file, dst+targetName)
	if err != nil {
		log.Println("[ERROR] SaveUploadedFile err: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}
	g.JSON(http.StatusOK, gin.H{
		"message": "Save file to " + dst + targetName,
	})
}

func PolicyGET(g *gin.Context) {
	g.JSON(http.StatusOK, gin.H{"messge": "OK", "policy": conf.VISUAL_NAVIGATION_MOD_POLICY})
}

func PolicyPOST(g *gin.Context) {
	policy := g.PostForm("policy")
	if policy == "UPLOAD" || policy == "SAVE" {
		conf.VISUAL_NAVIGATION_MOD_POLICY = policy
		g.JSON(http.StatusOK, gin.H{"messge": "OK"})
		return
	}
	g.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("policy %s not support", policy)})
}

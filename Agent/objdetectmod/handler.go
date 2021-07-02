package objdetectmod

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"simulator/Agent/conf"
	service "simulator/Agent/objdetectgrpc"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func Init(g *gin.Context) {
	addr := "0.0.0.0:" + conf.OBJ_DETECT_MOD_SERVICE_PORT
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Println("[ERROR] Can not connect to gRPC server: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}
	defer conn.Close()

	c := service.NewObjDetectModserviceClient(conn)
	r, err := c.Init(context.Background(),
		&service.InitArg{
			LabelPath:  "label_class.pbtxt",
			NumClasses: 5,
			QueueSize:  5,
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
	if conf.OBJ_DETECT_MOD_POLICY == "SAVE" || conf.ROLE == "CLOUD" {
		score, action, err = localInference(outputDirPath, sourceVideoPath, outputVideoPath)
		if err != nil {
			g.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	} else {
		action = "OBJECT CLOUD INFERENCE"
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
	targetName := key + "-objdetect" + ext
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
	g.JSON(http.StatusOK, gin.H{"messge": "OK", "policy": conf.OBJ_DETECT_MOD_POLICY})
}

func PolicyPOST(g *gin.Context) {
	policy := g.PostForm("policy")
	if policy == "UPLOAD" || policy == "SAVE" {
		conf.OBJ_DETECT_MOD_POLICY = policy
		g.JSON(http.StatusOK, gin.H{"messge": "OK"})
		return
	}
	g.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("policy %s not support", policy)})
}

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
	"simulator/Agent/model"
	service "simulator/Agent/objdetectgrpc"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// keep current tasknum
var TaskNum = 0

// Init AI model handler
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

// Inference task handler
func Inference(g *gin.Context) {
	dst := "./upload/"
	file, err := g.FormFile("file")
	if err != nil {
		log.Println("[ERROR] FormFile err: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}
	ext := filepath.Ext(file.Filename)            // .mp4
	key := strings.TrimSuffix(file.Filename, ext) // abc
	key = strings.Split(key, "-")[0]              // only keep hash value
	sourceName := key + "-objdetect-orig" + ext
	targetName := key + "-objdetect" + ext

	// file.Filename abc.mp4
	err = g.SaveUploadedFile(file, dst+sourceName)
	if err != nil {
		log.Println("[ERROR] SaveUploadedFile err: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	sourceVideoPath, _ := filepath.Abs(dst + sourceName)
	outputVideoPath, _ := filepath.Abs(dst + targetName)
	outputDirPath, _ := filepath.Abs(dst)

	var score float32
	var action string
	if conf.OBJ_DETECT_MOD_POLICY == model.POLICY_SAVE || conf.ROLE == model.ROLE_CLOUD {
		TaskNum += 1
		score, action, err = localInference(outputDirPath, sourceVideoPath, outputVideoPath)
		TaskNum -= 1
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

// Upload video handler
//
// back up valuable video
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
	key = strings.Split(key, "-")[0]              // only keep hash value
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

// Policy GET handler
//
// return current policy
func PolicyGET(g *gin.Context) {
	g.JSON(http.StatusOK, gin.H{"messge": "OK", "policy": conf.OBJ_DETECT_MOD_POLICY})
}

// Policy POST handler
//
// modify current policy
func PolicyPOST(g *gin.Context) {
	policy := g.PostForm("policy")
	if policy == model.POLICY_UPLOAD || policy == model.POLICY_SAVE {
		conf.OBJ_DETECT_MOD_POLICY = policy
		g.JSON(http.StatusOK, gin.H{"messge": "OK"})
		return
	}
	g.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("policy %s not support", policy)})
}

// Tasknum GET handler
//
// return current tasknum
func TaskNumGET(g *gin.Context) {
	g.JSON(http.StatusOK, gin.H{"message": "ok", "num": TaskNum})
}

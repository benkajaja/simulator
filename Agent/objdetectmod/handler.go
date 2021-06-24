package objdetectmod

import (
	"context"
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
	targetName := key + "-ok" + ext

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
	r, err := c.Inference(context.Background(),
		&service.InferenceArg{
			ModelPath: "model/frozen_inference_graph.pb",
			// VideoPath:              "Video/onesec.mp4",
			VideoPath:        dst + file.Filename,
			OverlapOutputDir: "overlap",
			OverlapOutput:    "no",
			DetectedOutput:   "yes",
			// DetectedOutputDir:      "output",
			// OutputLabeledVideoPath: "output/onesec.mp4",
			DetectedOutputDir:      dst,
			OutputLabeledVideoPath: dst + targetName,
		},
	)
	if err != nil {
		log.Println("[ERROR] Can not connect to gRPC server: ", err)
		g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}
	log.Println("[DEBUG] inference result: ", r.Status, r.Score)
	g.JSON(http.StatusOK, gin.H{
		"message": r.Status,
		"score":   r.Score,
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

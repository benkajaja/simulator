package main

import (
	"context"
	"log"
	"path/filepath"
	"sync"
	"time"

	objectgrpc "simulator/Agent/objdetectgrpc"
	visualgrpc "simulator/Agent/visualnavigationgrpc"

	"google.golang.org/grpc"
)

var TIMEWAIT = time.Duration(1)
var VIDEOLIST = []string{
	"186971ce281553fc.mp4",
	"34ed221a538be280.mp4",
	"4adacecb3fce47ef.mp4",
	"6d45e61a723f66ed.mp4",
	"7aad52bbf690d9e3.mp4",
	"8ba2ea4293039c30.mp4",
	"9306144e1e6252bb.mp4",
	"af63c74c8601c8dd.mp4",
	"ebe6a4f020a6a972.mp4",
	"f563464ab2b0ed43.mp4",
}
var VIDEODIR = "../Video/"
var LOCALURL = "http://localhost:8000"

type InferenceResp struct {
	Message string  `json:"message"`
	Score   float64 `json:"score"`
	Action  string  `json:"action"`
}

type UploadResp struct {
	Message string `json:"message"`
}

var concurrentTask = []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
var traceNum = []int{2, 2, 2, 3, 3, 4, 5, 7, 10, 20}

// var concurrentJob = map[int]int{
// 	10: 2,
// 	9:  2,
// 	8:  2,
// 	7:  3,
// 	6:  3,
// 	5:  4,
// 	4:  5,
// 	3:  7,
// 	2:  10,
// 	1:  20,
// }

func main() {
	for i := 0; i < len(concurrentTask); i++ {
		for j := 0; j < traceNum[i]; j++ {
			log.Printf("concurrentTask%d,trace#%d", concurrentTask[i], j)
			wg := new(sync.WaitGroup)
			wg.Add(concurrentTask[i])
			for c := 0; c < concurrentTask[i]; c++ {
				sourceVideoPath, _ := filepath.Abs("./Video/" + VIDEOLIST[c])
				outputVideoPath, _ := filepath.Abs("./upload/" + VIDEOLIST[c])
				go objdetectmodInference("./upload/", sourceVideoPath, outputVideoPath, concurrentTask[i], wg)
				// go visualnavigationmodInference(sourceVideoPath, outputVideoPath, concurrentTask[i], wg)
			}
			wg.Wait()
		}
	}
}

func objdetectmodInference(outputDirPath, sourceVideoPath, outputVideoPath string, concurrent int, wg *sync.WaitGroup) {
	defer wg.Done()
	t1 := time.Now()
	var err error
	addr := "0.0.0.0:" + "50052"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal("[ERROR] failed", err)
	}

	defer conn.Close()

	c := objectgrpc.NewObjDetectModserviceClient(conn)
	r, err := c.Inference(context.Background(),
		&objectgrpc.InferenceArg{
			ModelPath: "model/frozen_inference_graph.pb",
			// VideoPath:              "Video/onesec.mp4",
			VideoPath:        sourceVideoPath,
			OverlapOutputDir: "overlap",
			OverlapOutput:    "no",
			DetectedOutput:   "yes",
			// DetectedOutputDir:      "output",
			// OutputLabeledVideoPath: "output/onesec.mp4",
			DetectedOutputDir:      outputDirPath,
			OutputLabeledVideoPath: outputVideoPath,
		},
	)
	if err != nil {
		log.Fatal("[ERROR] failed", err)
	}
	t2 := time.Now()
	log.Printf("[DEBUG] OBJECT SCORE: %.2f TIME: %.3fs CONCURRENT: %d", r.Score, t2.Sub(t1).Seconds(), concurrent)
}
func visualnavigationmodInference(sourceVideoPath, outputVideoPath string, concurrent int, wg *sync.WaitGroup) {
	defer wg.Done()
	t1 := time.Now()
	var err error
	addr := "0.0.0.0:" + "50051"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal("[ERROR] failed", err)
	}

	defer conn.Close()
	c := visualgrpc.NewVisualNavigationModserviceClient(conn)
	r, err := c.Inference(context.Background(),
		&visualgrpc.InferenceArg{
			Input:  sourceVideoPath,
			Output: outputVideoPath,
		},
	)
	if err != nil {
		log.Fatal("[ERROR] failed", err)
	}
	t2 := time.Now()
	log.Printf("[DEBUG] VISUAL SCORE: %.2f TIME: %.3fs CONCURRENT: %d", r.Score, t2.Sub(t1).Seconds(), concurrent)
}

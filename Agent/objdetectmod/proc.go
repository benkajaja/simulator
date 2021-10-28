package objdetectmod

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"simulator/Agent/conf"
	"simulator/Agent/model"
	service "simulator/Agent/objdetectgrpc"
	"strings"
	"time"

	"google.golang.org/grpc"
)

type InferenceResp struct {
	Message string  `json:"message"`
	Score   float32 `json:"score"`
}

type UploadResp struct {
	Message string `json:"message"`
}

// local inference function
//
// include two step (inference, upload)
//
// Inference step: send video to local endpoint for inference via gRPC connection
//
// Upload step:
// If role == "CLOUD", no need to uplaod video for backing up.
// If role == "EDGE" and Score < uploadRatio, send video to cloud for backing up.
// 	Note: uploadRatio = 1 - uploadThreshold
func localInference(outputDirPath, sourceVideoPath, outputVideoPath string) (float32, string, error) {
	var score = float32(0)
	var action = "UNKNOWN"
	var err error

	t1 := time.Now()

	addr := "0.0.0.0:" + conf.OBJ_DETECT_MOD_SERVICE_PORT
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return score, action, err
	}

	defer conn.Close()

	c := service.NewObjDetectModserviceClient(conn)
	r, err := c.Inference(context.Background(),
		&service.InferenceArg{
			ModelPath: "model/frozen_inference_graph.pb",
			// VideoPath:              "Video/onesec.mp4",
			VideoPath:        sourceVideoPath,
			OverlapOutputDir: "overlap",
			OverlapOutput:    "no",
			DetectedOutput:   "no",
			// DetectedOutputDir:      "output",
			// OutputLabeledVideoPath: "output/onesec.mp4",
			DetectedOutputDir:      outputDirPath,
			OutputLabeledVideoPath: outputVideoPath,
		},
	)
	if err != nil {
		return score, action, err
	}
	log.Println("[DEBUG] inference result: ", r.Status, r.Score)

	filename := filepath.Base(sourceVideoPath)
	t2 := time.Now()
	filenameSplit := strings.Split(filename, "-")
	log.Printf("[DEBUG] OBJECT %5s %s INFERENCETIME: %.3fs", conf.ROLE, filenameSplit[0], t2.Sub(t1).Seconds())

	action = fmt.Sprintf("%s INFERENCE", conf.ROLE)

	if conf.ROLE == model.ROLE_CLOUD {
		action = "OBJECT CLOUD INFERENCE" // not used
		return r.Score, action, nil
	}

	if r.Score > conf.OBJ_DETECT_MOD_UPLOAD_RATIO {
		action = "OBJECT EDGE  INFERENCE NOT UPLOAD"
		return r.Score, action, nil
	}

	var uploadresp UploadResp
	status, uploadRawResp, err := sendPostReq(conf.CLOUDURL+"/objdetectmod/upload", sourceVideoPath, "file")
	if err != nil {
		return r.Score, action, err
	}
	if err := json.Unmarshal(uploadRawResp, &uploadresp); err != nil {
		return r.Score, action, errors.New("UNMARSHAL FAILED")
	}
	if status != http.StatusOK {
		return r.Score, action, errors.New("UPLOAD FAILED " + uploadresp.Message)
	}

	action = "OBJECT EDGE  INFERENCE UPLOAD"
	return r.Score, action, nil
}

// cloud inference function
//
// send video to cloud for inference
func cloudInference(sourceVideoPath string) (float32, error) {
	var score = float32(0)
	var resp InferenceResp
	status, inferenceRawResp, err := sendPostReq(conf.CLOUDURL+"/objdetectmod/inference", sourceVideoPath, "file")
	if err != nil {
		return score, err
	}
	if err := json.Unmarshal(inferenceRawResp, &resp); err != nil {
		return score, errors.New("UNMARSHAL FAILED")
	}
	if status != http.StatusOK {
		return score, errors.New("UPLOAD FAILED " + resp.Message)
	}
	return resp.Score, nil
}

// Send post request (contains video)
func sendPostReq(url, videopath, field string) (int, []byte, error) {
	file, err := os.Open(videopath)

	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(field, filepath.Base(file.Name()))

	if err != nil {
		return 0, nil, err
	}

	io.Copy(part, file)
	writer.Close()
	request, err := http.NewRequest("POST", url, body)

	if err != nil {
		return 0, nil, err
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return 0, nil, err
	}

	return response.StatusCode, content, nil
}

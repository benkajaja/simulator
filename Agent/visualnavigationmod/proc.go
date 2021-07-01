package visualnavigationmod

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"simulator/Agent/conf"
	service "simulator/Agent/visualnavigationgrpc"
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

func localInference(outputDirPath, sourceVideoPath, outputVideoPath string) (float32, string, error) {
	log.Println("[DEBUG] TAGA")
	var score = float32(0)
	var action = "UNKNOWN"
	var err error

	addr := "0.0.0.0:" + conf.VISUAL_NAVIGATION_MOD_SERVICE_PORT
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		// log.Println("[ERROR] Can not connect to gRPC server: ", err)
		// g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		// return
		return score, action, err
	}

	defer conn.Close()
	log.Println("[DEBUG] TAGB")
	c := service.NewVisualNavigationModserviceClient(conn)
	log.Println("[DEBUG] TAGC")
	r, err := c.Inference(context.Background(),
		&service.InferenceArg{
			Input:  sourceVideoPath,
			Output: outputVideoPath,
		},
	)
	if err != nil {
		// log.Println("[ERROR] Can not connect to gRPC server: ", err)
		// g.JSON(http.StatusInternalServerError, gin.H{"message": err})
		// return
		return score, action, err
	}
	log.Println("[DEBUG] inference result: ", r.Status, r.Score)

	if r.Score < conf.VISUAL_NAVIGATION_MOD_UPLOAD_THRESHOLD {
		// log.Println("[DEBUG]", mode, videopath, resp.Message, resp.Score, "NOT UPLOAD")
		action = "NOT UPLOAD"
		return r.Score, action, nil
	}

	var uploadresp UploadResp
	status, uploadRawResp, err := sendPostReq(conf.CLOUDURL+"/visualnavigationmod/upload", sourceVideoPath, "file")
	if err != nil {
		return r.Score, action, err
	}
	if err := json.Unmarshal(uploadRawResp, &uploadresp); err != nil {
		return r.Score, action, errors.New("UNMARSHAL FAILED")
	}
	if status != http.StatusOK {
		return r.Score, action, errors.New("UPLOAD FAILED " + uploadresp.Message)
	}

	action = "UPLOAD"
	return r.Score, action, nil
}

func cloudInference(sourceVideoPath string) (float32, error) {
	var score = float32(0)
	var resp InferenceResp
	status, inferenceRawResp, err := sendPostReq(conf.CLOUDURL+"/visualnavigationmod/inference", sourceVideoPath, "file")
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
	// log.Println(string(content))
}

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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
)

type InferenceResp struct {
	Message string  `json:"message"`
	Score   float32 `json:"score"`
}

type UploadResp struct {
	Message string `json:"message"`
}

func localInference(ctx context.Context, outputDirPath, sourceVideoPath, outputVideoPath string) (float32, string, error) {
	tr := otel.Tracer("")
	_, span := tr.Start(ctx, "localinference")
	defer span.End()
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
	c := service.NewVisualNavigationModserviceClient(conn)
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

	switch conf.ROLE {
	case "CLOUD":
		action = "VISUAL CLOUD INFERENCE"
		return r.Score, action, nil
	case "EDGE":
		action = "VISUAL EDGE  INFERENCE"
		return r.Score, action, nil
	}
	return r.Score, action, nil
}

func cloudInference(ctx context.Context, sourceVideoPath string) (float32, error) {
	tr := otel.Tracer("")
	_, span := tr.Start(ctx, "cloudinference")
	defer span.End()
	var score = float32(0)
	var resp InferenceResp
	status, inferenceRawResp, err := sendPostReq(ctx, conf.CLOUDURL+"/visualnavigationmod/inference", sourceVideoPath, "file")
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

func sendPostReq(ctx context.Context, url, videopath, field string) (int, []byte, error) {
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
	// var hc propagation.HeaderCarrier
	// hc = request.Header
	// var hc propagation.HeaderCarrier{
	// 	Header: request.Header
	// }
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(request.Header))
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

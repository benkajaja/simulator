package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
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
var VIDEODIR = "./Video"
var CLOUDURL = "http://***REMOVED***:8000"
var LOCALURL = "http://localhost:8000"
var UPLOADTHRESHOLD = 0.5

type InferenceResp struct {
	Message string  `json:"message"`
	Score   float64 `json:"score"`
}

type UploadResp struct {
	Message string `json:"message"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	wg := new(sync.WaitGroup)
	wg.Add(len(VIDEOLIST))

	for _, v := range VIDEOLIST {
		videopath := fmt.Sprintf("%s/%s", VIDEODIR, v)
		mode := GetMode()
		go SendInferenceRequest(mode, videopath, wg)
		log.Println("[DEBUG] INFERENCE", v)
		time.Sleep(TIMEWAIT * time.Second)
	}
	wg.Wait()

	log.Println("[DEBUG] All done")
}

func GetMode() string {
	var mode string
	switch rand.Float64() > 0.5 {
	case (true):
		mode = "CLOUD"
	case (false):
		mode = "LOCAL"
	}
	return mode
}

func SendInferenceRequest(mode string, videopath string, wg *sync.WaitGroup) {
	defer wg.Done()
	var url string
	switch mode {
	case "CLOUD":
		url = CLOUDURL
	case "LOCAL":
		url = LOCALURL
	default:
		url = LOCALURL
	}

	var resp InferenceResp
	inferenceRawResp := SendPostReq(url+"/objdetectmod/inference", videopath, "file")
	if err := json.Unmarshal(inferenceRawResp, &resp); err != nil {
		log.Fatal(err)
	}
	// log.Println("[DEBUG]", mode, resp.Message, resp.Score)

	if mode == "CLOUD" {
		log.Println("[DEBUG]", mode, videopath, resp.Message, resp.Score)
		return
	}

	if resp.Score < UPLOADTHRESHOLD {
		log.Println("[DEBUG]", mode, videopath, resp.Message, resp.Score, "NOT UPLOAD")
		return
	}
	var uploadresp UploadResp
	uploadRawResp := SendPostReq(CLOUDURL+"/objdetectmod/upload", videopath, "file")
	if err := json.Unmarshal(uploadRawResp, &uploadresp); err != nil {
		log.Fatal(err)
	}
	log.Println("[DEBUG]", mode, videopath, resp.Message, resp.Score, "DO UPLOAD")
}

func SendPostReq(url, videopath, field string) []byte {
	file, err := os.Open(videopath)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(field, filepath.Base(file.Name()))

	if err != nil {
		log.Fatal(err)
	}

	io.Copy(part, file)
	writer.Close()
	request, err := http.NewRequest("POST", url, body)

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}
	return content
	// log.Println(string(content))
}

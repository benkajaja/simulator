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

var TIMEWAIT = time.Duration(1 * time.Second)
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
var LOCALURL = "http://localhost:8000"

type InferenceResp struct {
	Message string  `json:"message"`
	Score   float64 `json:"score"`
	Action  string  `json:"action"`
}

type UploadResp struct {
	Message string `json:"message"`
}

func main() {
	t1 := time.Now()
	rand.Seed(time.Now().UnixNano())

	wg := new(sync.WaitGroup)
	wg.Add(2 * len(VIDEOLIST))

	for k, v := range VIDEOLIST {
		videopath := fmt.Sprintf("%s/%s", VIDEODIR, v)
		go SendInferenceRequest("objdetectmod", videopath, wg)
		go SendInferenceRequest("visualnavigationmod", videopath, wg)
		log.Println("[DEBUG] INFERENCE", k, v)
		time.Sleep(TIMEWAIT)
	}
	wg.Wait()
	t2 := time.Now()
	log.Printf("[DEBUG] All done %.2fs", t2.Sub(t1).Seconds())
}

func SendInferenceRequest(service string, videopath string, wg *sync.WaitGroup) {
	t1 := time.Now()
	defer wg.Done()
	var url = LOCALURL

	var resp InferenceResp
	var status int
	serviceURL := fmt.Sprintf("%s/%s/inference", url, service) //objdetectmod
	status, inferenceRawResp := SendPostReq(serviceURL, videopath, "file")
	if err := json.Unmarshal(inferenceRawResp, &resp); err != nil {
		log.Fatal(err)
	}
	if status != http.StatusOK {
		log.Fatal(resp.Message)
	}
	t2 := time.Now()
	log.Printf("[DEBUG] %s %.2f %s %.2fs", filepath.Base(videopath), resp.Score, resp.Action, t2.Sub(t1).Seconds())

}

func SendPostReq(url, videopath, field string) (int, []byte) {
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
	return response.StatusCode, content
}

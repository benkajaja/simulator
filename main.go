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
	"f563464ab2b0ed43.mp4", //10

	"a91b2f479c78380a.mp4",
	"cbf33b454853d968.mp4",
	"b420580c0e8004cf.mp4",
	"b5d3d49589157b2c.mp4",
	"41eafc7cd9083b59.mp4",
	"e4fd5098d8c6c2f2.mp4",
	"c2fb61cd102c8dc4.mp4",
	"f053acbb14d7f8b8.mp4",
	"59064c080200000c.mp4",
	"f39a2c02233fdcbc.mp4", //20

	"f033cf6ccc08c1ba.mp4",
	"108930c7f004595c.mp4",
	"fa81d82da1e60cc7.mp4",
	"e328210ee6eec0d4.mp4",
	"7a882dffd7f66e5d.mp4",
	"5f9b3baef2ddf136.mp4",
	"4cdaffb90d22b3d7.mp4",
	"ef2d347d1f94fd3b.mp4",
	"b2e00ea7bb8f523b.mp4",
	"13ec628ed487a191.mp4", //30

	"4bb6493dff0a1d14.mp4",
	"504fb0fbd42e89ae.mp4",
	"cdd327bb30b081af.mp4",
	"f02e88cf140a2013.mp4",
	"e59e1894001c58aa.mp4",
	"972178f46c480a7a.mp4",
	"694b7876b48a810a.mp4",
	"34a92fa03e95c8f9.mp4",
	"fdbcdb3518741c6b.mp4",
	"cd7f02f760c6e113.mp4", //40
}
var VIDEODIR = "./Video/1s"
var LOCALURL = "http://localhost:8000"

type InferenceResp struct {
	Message string  `json:"message"`
	Score   float64 `json:"score"`
	Action  string  `json:"action"`
}

type UploadResp struct {
	Message string `json:"message"`
}

// Send video in VIDEOLIST to objdetectmod and visualnavigationmod
//
//  Send 1-second slice: VIDEODIR = "./Video/1s", len(VIDEOLIST) = 40, TIMEWAIT = 1
//  Send 4-second slice: VIDEODIR = "./Video/4s", len(VIDEOLIST) = 10, TIMEWAIT = 4
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
	serviceURL := fmt.Sprintf("%s/%s/inference", url, service)
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

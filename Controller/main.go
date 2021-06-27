package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"simulator/Controller/conf"
	"syscall"
	"time"
)

type StatusResp struct {
	GPUUtil    float32 `json:"gpuUtil"`
	GPUMemUtil float32 `json:"gpumemUtil"`
}

type PolicyGETResp struct {
	Message string `json:"message"`
	Policy  string `json:"policy"`
}

type PolicyPOSTResp struct {
	Message string `json:"message"`
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			test()
		case <-sigs:
			return
		}
	}
}

func test() {
	host := conf.EDGELIST[0]
	resp, err := http.Get("http://" + host + "/status")
	if err != nil {
		// handle error
		log.Println("[ERROR] failed to get", host, err)
		return
	}

	defer resp.Body.Close()

	var data StatusResp
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println("[ERROR] failed to decode", err)
		return
	}

	var policy string
	switch {
	case data.GPUUtil > conf.CHANGE_POLICY_GPUUTIL_THRESHOLD:
		policy = "UPLOAD"
	default:
		policy = "SAVE"
	}
	if err := changePolicy(host, "objdetectmod", policy); err != nil {
		log.Printf("[ERROR] changePolicy %s %s %s failed %s", host, "objdetectmod", policy, err.Error())
	}

	log.Printf("[DEBUG] GPU:%.2f GPUMEM:%.2f POLICY:%s", data.GPUUtil, data.GPUMemUtil, policy)
}

func changePolicy(host, service, policy string) error {
	var err error
	posturl := fmt.Sprintf("http://%s/%s/policy", host, service)
	data := url.Values{
		"policy": {policy},
	}
	resp, err := http.PostForm(posturl, data)
	if err != nil {
		return err
	}

	var postResp PolicyPOSTResp
	if err = json.NewDecoder(resp.Body).Decode(&postResp); err != nil {
		return errors.New("decode FAILED")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(postResp.Message)
	}
	return nil
}

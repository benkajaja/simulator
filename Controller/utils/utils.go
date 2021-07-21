package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"simulator/Controller/conf"
	"simulator/Controller/model"
	"time"
)

type TaskNumGETResp struct {
	Message string `json:"message"`
	Num     int    `json:"num"`
}

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

func ProbeBW(url string) (float32, error) {
	var err error
	var res float32 // in Mbps

	t1 := time.Now()
	resp, err := http.Get(fmt.Sprintf("http://%s/probe", url))
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0.0, err
	}
	t2 := time.Now()
	res = float32(conf.PROBE_FILE_SIZE_KB*8) / float32(t2.Sub(t1).Seconds()*1024)
	return res, nil
}

func GetConcurrentTask(url, service string) (int, error) {
	var err error
	var res int

	resp, err := http.Get(fmt.Sprintf("http://%s/%s/tasknum", url, service))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data TaskNumGETResp
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	res = data.Num
	return res, nil
}

func Update() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for _, e := range conf.EDGELIST {
		if !e.IsProbing && e.ProbeCoolDownCounter <= 0 {
			go func(e *model.EdgeNode, ctx context.Context) {
				e.IsProbing = true
				res, err := ProbeBW(e.Url)
				if err != nil {
					log.Printf("[ERROR] failed to probe node %s", e.Url)
					return
				}
				e.EstBW = res
				if float32(conf.PROBE_FILE_SIZE_KB*8)/(res*1024) > 1 {
					e.ProbeCoolDownCounter = 20
				}
				e.ProbDuration = 0
				e.IsProbing = false
			}(e, ctx)
		} else if e.IsProbing {
			e.ProbDuration += 1
			e.EstBW = e.EstBW * float32(e.ProbDuration) / float32(e.ProbDuration+1)
		} else if e.ProbeCoolDownCounter > 0 {
			e.ProbeCoolDownCounter -= 1
		}

		for _, j := range e.Services.(map[string]interface{}) {
			go func(s *model.Service) {
				res, err := GetConcurrentTask(e.Url, s.ServiceUrl)
				if err != nil {
					log.Printf("[ERROR] failed to get tasknum %s %s", e.Url, s.ServiceUrl)
					return
				}
				s.ConcurrentTask = res
			}(j.(*model.Service))
		}
	}

	for _, c := range conf.CLOUDLIST {
		for _, j := range c.Services.(map[string]interface{}) {
			go func(s *model.Service) {
				res, err := GetConcurrentTask(c.Url, s.ServiceUrl)
				if err != nil {
					log.Printf("[ERROR] failed to get tasknum %s %s", c.Url, s.ServiceUrl)
					return
				}
				s.ConcurrentTask = res
			}(j.(*model.Service))
		}
	}
}

func ChangePolicy(host, service, policy string) error {
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

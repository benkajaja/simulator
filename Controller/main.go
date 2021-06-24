package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"simulator/Controller/conf"
	"syscall"
	"time"
)

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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		log.Println("[ERROR] failed to read body")
		return
	}

	fmt.Println(string(body))

}

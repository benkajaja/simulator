package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"simulator/Controller/conf"
	"simulator/Controller/core"
	"simulator/Controller/model"
	"simulator/Controller/utils"
	"syscall"
	"time"
)

func main() {
	if err := conf.Init("./conf.json"); err != nil {
		log.Fatal("[ERROR] load conf.json fail", err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Trigger update, gandancore, debugmessage every seconds
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go utils.Update()
			go core.GanDanCore()
			go debugmessage()
		case <-sigs:
			return
		}
	}
}

func debugmessage() {
	for i := range conf.EDGELIST {
		fmt.Printf("%s %s %f %d\n", i, conf.EDGELIST[i].Url, conf.EDGELIST[i].EstBW, conf.EDGELIST[i].ProbeCoolDownCounter)
		for j, k := range conf.EDGELIST[i].Services.(map[string]interface{}) {
			fmt.Printf("%s %d\n", j, k.(*model.Service).ConcurrentTask)
		}
	}
	for i := range conf.CLOUDLIST {
		fmt.Printf("%s %s\n", i, conf.CLOUDLIST[i].Url)
		for j, k := range conf.CLOUDLIST[i].Services.(map[string]interface{}) {
			fmt.Printf("%s %d\n", j, k.(*model.Service).ConcurrentTask)
		}
	}
}

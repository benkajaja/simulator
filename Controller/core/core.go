package core

import (
	"simulator/Controller/algo"
	"simulator/Controller/conf"
	"simulator/Controller/model"
	"simulator/Controller/utils"
)

func GanDanCore() {
	targetCloud := conf.CLOUDLIST["local"]
	cloudsvc := targetCloud.Services.(map[string]interface{})
	for _, e := range conf.EDGELIST {
		for sname, s := range e.Services.(map[string]interface{}) {
			go func(sname string, s *model.Service) {
				policy := algo.TaroAlgo(e, s, cloudsvc[sname].(*model.Service))
				if cloudsvc[sname].(*model.Service).Policy != policy {
					cloudsvc[sname].(*model.Service).Policy = policy
					utils.ChangePolicy(e.Url, s.ServiceUrl, policy)
				}
			}(sname, s.(*model.Service))
		}
	}
}

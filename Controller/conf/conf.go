package conf

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"simulator/Controller/model"

	"github.com/goinggo/mapstructure"
)

var CONTROLLER_PORT string
var CHANGE_POLICY_GPUUTIL_THRESHOLD float32
var EDGELIST = make(map[string]*model.EdgeNode)
var PROBE_FILE_SIZE_KB int

func Init(confpath string) error {
	var res map[string]interface{}
	var controllerconf map[string]interface{}
	var err error

	f, _ := os.Open(confpath)
	defer f.Close()

	b, _ := ioutil.ReadAll(f)
	if err = json.Unmarshal(b, &res); err != nil {
		return err
	}

	controllerconf = res["controller"].(map[string]interface{})
	CONTROLLER_PORT = controllerconf["CONTROLLER_PORT"].(string)
	CHANGE_POLICY_GPUUTIL_THRESHOLD = float32(controllerconf["CHANGE_POLICY_GPUUTIL_THRESHOLD"].(float64))
	PROBE_FILE_SIZE_KB = int(controllerconf["PROBE_FILE_SIZE_KB"].(float64))

	s := controllerconf["EDGELIST"].(map[string]interface{})
	for k, v := range s {
		var e model.EdgeNode
		if err = mapstructure.Decode(v, &e); err != nil {
			return err
		}
		for i, j := range e.Services.(map[string]interface{}) {
			var s model.Service
			log.Println(i, j)
			if err = mapstructure.Decode(j, &s); err != nil {
				return err
			}
			e.Services.(map[string]interface{})[i] = &s
		}
		EDGELIST[k] = &e
		log.Println(e)
	}
	return nil
}

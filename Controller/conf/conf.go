package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var CONTROLLER_PORT string
var CHANGE_POLICY_GPUUTIL_THRESHOLD float32
var EDGELIST []string

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
	s := controllerconf["EDGELIST"].([]interface{})
	for _, v := range s {
		EDGELIST = append(EDGELIST, v.(string))
	}
	return nil
}

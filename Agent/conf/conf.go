package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var OBJ_DETECT_MOD_SERVICE_PORT string
var OBJ_DETECT_MOD_POLICY string // {UPLOAD, SAVE}
var OBJ_DETECT_MOD_UPLOAD_THRESHOLD float32

var VISUAL_NAVIGATION_MOD_SERVICE_PORT string
var VISUAL_NAVIGATION_MOD_POLICY string // {UPLOAD, SAVE}

var CLOUDURL string
var AGENT_PORT string
var ROLE string // {CLOUD, EDGE}

func Init(confpath string) error {
	var res map[string]interface{}
	var agentconf map[string]interface{}
	var err error

	f, _ := os.Open(confpath)
	defer f.Close()

	b, _ := ioutil.ReadAll(f)
	if err = json.Unmarshal(b, &res); err != nil {
		return err
	}

	agentconf = res["agent"].(map[string]interface{})
	OBJ_DETECT_MOD_SERVICE_PORT = agentconf["OBJ_DETECT_MOD_SERVICE_PORT"].(string)
	OBJ_DETECT_MOD_POLICY = agentconf["OBJ_DETECT_MOD_POLICY"].(string)
	OBJ_DETECT_MOD_UPLOAD_THRESHOLD = float32(agentconf["OBJ_DETECT_MOD_UPLOAD_THRESHOLD"].(float64))
	VISUAL_NAVIGATION_MOD_SERVICE_PORT = agentconf["VISUAL_NAVIGATION_MOD_SERVICE_PORT"].(string)
	VISUAL_NAVIGATION_MOD_POLICY = agentconf["VISUAL_NAVIGATION_MOD_POLICY"].(string)
	CLOUDURL = agentconf["CLOUDURL"].(string)
	AGENT_PORT = agentconf["AGENT_PORT"].(string)
	ROLE = agentconf["ROLE"].(string)

	return nil
}

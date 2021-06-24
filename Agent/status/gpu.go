package status

import (
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

func gpuinfo() (float32, float32, error) {
	var err error
	nvml.Init()
	defer nvml.Shutdown()
	device, err := nvml.NewDevice(0)
	if err != nil {
		return 0, 0, err
	}
	st, err := device.Status()
	if err != nil {
		return 0, 0, err
	}
	return float32(*st.Utilization.GPU) / 100.0,
		float32(*st.Memory.Global.Used) / float32(*device.Memory),
		nil
}

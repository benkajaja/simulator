package algo

import (
	"simulator/Controller/conf"
	"simulator/Controller/model"
)

func TaroAlgo(e *model.EdgeNode, se *model.Service, sc *model.Service) string {
	// est completion time on edge
	estInferenceTimeE := G(se.ConcurrentTask, se.InferenceTimeSlope, se.InferenceTimeConstant)
	estTransferTimeE := se.UploadProb * F(conf.VIDEO_FILE_SIZE_KB*1024*8, e.EstBW)

	// est completion time on cloud
	estInferenceTimeC := G(sc.ConcurrentTask, sc.InferenceTimeSlope, sc.InferenceTimeConstant)
	estTransferTimeC := 1.0 * F(conf.VIDEO_FILE_SIZE_KB*1024*8, e.EstBW)

	if estInferenceTimeE+estTransferTimeE < estInferenceTimeC+estTransferTimeC {
		return "SAVE"
	} else {
		return "UPLOAD"
	}
}

// return est. inference time
func G(taskNum int, slope, constant float32) float32 {
	return float32(taskNum)*slope + constant
}

// fileSize in bits,
// bw in Mbps,
// return est. transfer time
func F(fileSize int, bw float32) float32 {
	return float32(fileSize) / (bw * 1024 * 1024)
}

package model

type Service struct {
	ServiceUrl            string
	ConcurrentTask        int
	InferenceTimeSlope    float32
	InferenceTimeConstant float32 // est. inference time = task#*slope + constant
}

type EdgeNode struct {
	Url      string
	EstBW    float32 // in Mbps
	Services interface{}
}

type CloudNode struct {
	Services interface{}
}

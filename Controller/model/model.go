package model

type Service struct {
	ServiceUrl            string
	UploadProb            float32
	ConcurrentTask        int
	InferenceTimeSlope    float32
	InferenceTimeConstant float32 // est. inference time = task#*slope + constant
}

type EdgeNode struct {
	Url                  string
	EstBW                float32 // in Mbps
	ProbeCoolDownCounter int
	IsProbing            bool
	ProbDuration         int // in second
	Services             interface{}
}

type CloudNode struct {
	Url      string
	Services interface{}
}

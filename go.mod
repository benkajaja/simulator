module simulator

go 1.16

require (
	github.com/NVIDIA/gpu-monitoring-tools v0.0.0-20210623151644-b29849604a14
	github.com/NVIDIA/gpu-monitoring-tools/bindings/go/dcgm v0.0.0-20210525204842-7fbe8db26700
	github.com/gin-gonic/gin v1.7.2
	github.com/goinggo/mapstructure v0.0.0-20140717182941-194205d9b4a9
	go.opentelemetry.io/otel v1.0.0-RC2
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.0.0-RC2
	go.opentelemetry.io/otel/sdk v1.0.0-RC2
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)

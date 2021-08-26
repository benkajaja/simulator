package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"simulator/Agent/conf"
	"simulator/Agent/objdetectmod"
	"simulator/Agent/status"
	"simulator/Agent/visualnavigationmod"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var exporter tracesdk.SpanExporter
var tp *tracesdk.TracerProvider

func initTracer() {
	var err error
	// exporter, err = jaeger.New(
	// 	jaeger.WithAgentEndpoint(
	// 		jaeger.WithAgentHost(""),
	// 		jaeger.WithAgentPort("6831"),
	// 	),
	// )

	exporter, err = stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	initTracer()
	rand.Seed(time.Now().UnixNano())
	if err := conf.Init("./conf.json"); err != nil {
		log.Fatal("[ERROR] load conf.json fail", err)
	}
	tp = tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(
			// resource.NewWithAttributes(
			// 	semconv.SchemaURL,
			// 	semconv.ServiceNameKey.String("haha"),
			// 	// attribute.String("environment", environment),
			// 	// attribute.Int64("ID", id),
			// ),
			resource.NewSchemaless(
				semconv.ServiceNameKey.String("haha"),
			),
		),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	r := gin.Default()
	// r.Use(otelgin.Middleware(
	// 	"exporter",
	// 	otelgin.WithTracerProvider(
	// 		tp,
	// 	),
	// ))
	r.GET("status", status.Statuscheck)
	r.GET("probe", status.Probe)
	objDetectModService := r.Group("/objdetectmod")
	objDetectModService.GET("/init", objdetectmod.Init)
	objDetectModService.POST("/inference", objdetectmod.Inference)
	objDetectModService.POST("/upload", objdetectmod.Upload)
	objDetectModService.GET("/policy", objdetectmod.PolicyGET)
	objDetectModService.POST("/policy", objdetectmod.PolicyPOST)
	objDetectModService.GET("/tasknum", objdetectmod.TaskNumGET)

	visualNavigationModService := r.Group("/visualnavigationmod")
	visualNavigationModService.GET("/init", visualnavigationmod.Init)
	visualNavigationModService.POST("/inference", visualnavigationmod.Inference)
	visualNavigationModService.POST("/upload", visualnavigationmod.Upload)
	visualNavigationModService.GET("/policy", visualnavigationmod.PolicyGET)
	visualNavigationModService.POST("/policy", visualnavigationmod.PolicyPOST)
	visualNavigationModService.GET("/tasknum", visualnavigationmod.TaskNumGET)

	r.GET("/test", testtracer)
	r.Run("0.0.0.0:" + conf.AGENT_PORT)
}

func testtracer(c *gin.Context) {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	tr := tp.Tracer("component-main")
	cctx, span := tr.Start(c.Request.Context(), "foo")
	// var hc propagation.HeaderCarrier
	// hc = c.Request.Header
	// otel.GetTextMapPropagator().Inject(cctx, hc)
	defer span.End()
	subfunction(cctx)
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func subfunction(ctx context.Context) {
	// tr := otel.Tracer("component-bar")
	// _, span := tr.Start(context.Background(), "bar", trace.WithSpanKind(trace.SpanKindInternal))
	// span.SetAttributes(attribute.Key("testset").String("value"))
	// defer span.End()

	tr := otel.Tracer("component-bar")
	_, span := tr.Start(ctx, "bar")
	span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()
	log.Println("hello")
	time.Sleep(2 * time.Second)
}

package tracing

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/setherplatform/sether-node/tracing"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"gopkg.in/urfave/cli.v1"
)

var EnableFlag = cli.BoolFlag{
	Name:  "tracing",
	Usage: "Enable traces collection and reporting",
}

func Start(ctx *cli.Context) (stop func(), err error) {
	stop = func() {}

	if !ctx.Bool(EnableFlag.Name) {
		return
	}

	var cfg *jaegercfg.Configuration
	cfg, err = jaegercfg.FromEnv()
	if err != nil {
		return
	}

	cfg.ServiceName = "sether"

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return
	}
	stop = func() {
		closer.Close()
	}

	opentracing.SetGlobalTracer(tracer)
	tracing.SetEnabled(true)
	return
}

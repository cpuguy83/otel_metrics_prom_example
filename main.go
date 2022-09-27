package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	ctxB := context.Background()
	ctx, cancel := signal.NotifyContext(ctxB, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	exp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure(), otlpmetrichttp.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")))
	if err != nil {
		panic(err)
	}

	rdr := sdkmetric.NewPeriodicReader(exp)
	p := sdkmetric.NewMeterProvider(sdkmetric.WithReader(rdr))
	defer func() {
		fmt.Fprintln(os.Stderr, "collecting final metrics...")
		m, err := rdr.Collect(ctxB)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to collect metrics:", err)
			return
		}
		if err := exp.Export(ctxB, m); err != nil {
			fmt.Fprintln(os.Stderr, "failed to export metrics:", err)
		}
		if err := p.Shutdown(ctxB); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	global.SetMeterProvider(p)

	doStuff(ctx)
}

func doStuff(ctx context.Context) {
	p := global.MeterProvider()
	ctr, err := p.Meter("test").SyncInt64().Counter("testdata.counter", instrument.WithDescription("test counter"), instrument.WithUnit("1"))
	if err != nil {
		panic(err)
	}

	ctr.Add(ctx, 1, attribute.String("foo", "bar"))
}

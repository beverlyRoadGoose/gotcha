package monitoring

import (
	"context"
	"errors"
	"time"

	"tobi.ad/gotcha/config"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/shirou/gopsutil/v3/mem"
)

var meter = otel.Meter("tobi.ad/gotcha/monitoring")

func InitialiseOtelMonitoring(ctx context.Context, config *config.Monitoring) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// Shutdown calls cleanupBeforeShutdown functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanupBeforeShutdown will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanupBeforeShutdown and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	otel.SetTextMapPropagator(newPropagator())

	if config.Tracing.Enabled {
		traceProvider, innerErr := newTraceProvider(ctx, config.Tracing.SampleRate)
		if innerErr != nil {
			handleErr(innerErr)
			return
		}
		otel.SetTracerProvider(traceProvider)
	}

	if config.LogsEnabled {
		// Set up a logger provider.
		loggerProvider, err := newLoggerProvider(ctx)
		if err != nil {
			handleErr(err)
			return
		}
		shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
		global.SetLoggerProvider(loggerProvider)
	}

	if config.MetricsEnabled {
		// Set up meter provider.
		meterProvider, err := newMeterProvider(ctx)
		if err != nil {
			handleErr(err)
			return
		}
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)

		// enable the runtime metrics
		err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
		if err != nil {
			handleErr(err)
			return
		}

		err = exportSystemMemory()
		if err != nil {
			handleErr(err)
			return
		}
	}

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, sampleRate float64) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter, trace.WithBatchTimeout(5*time.Second)),
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(sampleRate))), // sample 10%
	)
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(15*time.Second))),
	)

	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context) (*log.LoggerProvider, error) {
	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)

	return loggerProvider, nil
}

func exportSystemMemory() error {
	memTotal, err := meter.Int64ObservableGauge("system.memory.total")
	if err != nil {
		return err
	}

	memUsed, err := meter.Int64ObservableGauge("system.memory.used")
	if err != nil {
		return err
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, observer otelmetric.Observer) error {
		v, err := mem.VirtualMemory()
		if err != nil {
			return err
		}

		observer.ObserveInt64(memTotal, int64(v.Total))
		observer.ObserveInt64(memUsed, int64(v.Used))

		return nil
	}, memTotal, memUsed)

	return err
}

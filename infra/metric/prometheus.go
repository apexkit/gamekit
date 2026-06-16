package metric

import (
	"net/http"
	"sync"

	"github.com/go-kratos/kratos/v2/middleware/metrics"
	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type KratosMeter struct {
	Requests otelmetric.Int64Counter
	Seconds  otelmetric.Float64Histogram
}

var (
	setupOnce sync.Once
	setupErr  error
)

func SetupPrometheusMeterProvider() error {
	setupOnce.Do(func() {
		exporter, err := otelprom.New()
		if err != nil {
			setupErr = err
			return
		}

		collectors.NewGoCollector()
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})

		provider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(exporter),
		)
		otel.SetMeterProvider(provider)
	})

	return setupErr
}

func MustSetupPrometheusMeterProvider() {
	if err := SetupPrometheusMeterProvider(); err != nil {
		logrus.Panic(err)
	}
}

func NewKratosMeter(jobName string) (*KratosMeter, error) {
	if err := SetupPrometheusMeterProvider(); err != nil {
		return nil, err
	}

	meter := otel.Meter(jobName)
	requests, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		return nil, err
	}

	seconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		return nil, err
	}

	return &KratosMeter{
		Requests: requests,
		Seconds:  seconds,
	}, nil
}

func MustNewKratosMeter(jobName string) *KratosMeter {
	meter, err := NewKratosMeter(jobName)
	if err != nil {
		logrus.Panic(err)
	}
	return meter
}

func Handler() http.Handler {
	return promhttp.Handler()
}

func Gatherer() promclient.Gatherer {
	return promclient.DefaultGatherer
}

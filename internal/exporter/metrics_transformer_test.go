package exporter

import (
	"testing"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// TestTransformHistogramWithNoAttributes reproduce el panic original:
// "assignment to entry in nil map" cuando un histogram data point no tiene atributos.
func TestTransformHistogramWithNoAttributes(t *testing.T) {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "test-service")

	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName("http.server.duration")
	metric.SetEmptyHistogram()

	dp := metric.Histogram().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetCount(10)
	dp.SetSum(500.0)
	dp.ExplicitBounds().FromRaw([]float64{0.005, 0.01, 0.025, 0.05, 0.1})
	dp.BucketCounts().FromRaw([]uint64{1, 2, 3, 2, 1, 1})
	// Sin atributos en el data point — esto causaba el panic

	records, err := TransformMetrics(md)
	if err != nil {
		t.Fatalf("TransformMetrics returned error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].MetricType != "histogram" {
		t.Errorf("expected MetricType 'histogram', got %s", records[0].MetricType)
	}
	if records[0].ServiceName != "test-service" {
		t.Errorf("expected ServiceName 'test-service', got %s", records[0].ServiceName)
	}
}

// TestTransformExponentialHistogramWithNoAttributes verifica el mismo bug
// en exponential histograms.
func TestTransformExponentialHistogramWithNoAttributes(t *testing.T) {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "test-service")

	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName("http.server.duration.exp")
	metric.SetEmptyExponentialHistogram()

	dp := metric.ExponentialHistogram().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetCount(5)
	dp.SetSum(250.0)
	dp.SetScale(1)
	// Sin atributos en el data point

	records, err := TransformMetrics(md)
	if err != nil {
		t.Fatalf("TransformMetrics returned error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].MetricType != "exponential_histogram" {
		t.Errorf("expected MetricType 'exponential_histogram', got %s", records[0].MetricType)
	}
}

// TestTransformSummaryWithNoAttributes verifica el mismo bug en summaries.
func TestTransformSummaryWithNoAttributes(t *testing.T) {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "test-service")

	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName("process.runtime.gc.duration")
	metric.SetEmptySummary()

	dp := metric.Summary().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetCount(100)
	dp.SetSum(1000.0)
	qv := dp.QuantileValues().AppendEmpty()
	qv.SetQuantile(0.99)
	qv.SetValue(9.5)
	// Sin atributos en el data point

	records, err := TransformMetrics(md)
	if err != nil {
		t.Fatalf("TransformMetrics returned error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].MetricType != "summary" {
		t.Errorf("expected MetricType 'summary', got %s", records[0].MetricType)
	}
}

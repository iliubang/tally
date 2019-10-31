package prometheus

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	prom "github.com/m3db/client_golang/prometheus"
	model "github.com/m3db/prometheus_client_model/go"
)

type jsonTransformer struct {
	gatherer prom.Gatherer
}

func newJsonTransformer(g prom.Gatherer) *jsonTransformer {
	return &jsonTransformer{
		gatherer: g,
	}
}

type generic interface{}

func (t *jsonTransformer) toJson() []byte {
	result := make(map[string]generic)
	gathering, _ := t.gatherer.Gather()
	for _, mf := range gathering {
		switch mf.GetType() {
		case model.MetricType_COUNTER:
			result[mf.GetName()] = t.exportCounter(mf.GetMetric())
		case model.MetricType_GAUGE:
			result[mf.GetName()] = t.exportGauge(mf.GetMetric())
		case model.MetricType_HISTOGRAM:
			result[mf.GetName()] = t.exportHistogram(mf.GetMetric())
		case model.MetricType_SUMMARY:
			result[mf.GetName()] = t.exportSummary(mf.GetMetric())
		}
	}
	j, _ := json.Marshal(result)
	return j
}

func (t *jsonTransformer) exportLables(labels []*model.LabelPair) generic {
	res := make(map[string]generic, len(labels))
	for _, label := range labels {
		res[label.GetName()] = label.GetValue()
	}
	return res
}

func (t *jsonTransformer) exportHistogram(metrics []*model.Metric) generic {
	res := make([]generic, len(metrics))
	for idx, m := range metrics {
		val := make(map[string]generic)
		h := m.GetHistogram()
		lables := m.GetLabel()
		val["Type"] = "Histogram"
		val["Tags"] = t.exportLables(lables)
		val["Count"] = h.GetSampleCount()
		if math.IsNaN(h.GetSampleSum()) {
			val["Sum"] = 0.0
		} else {
			val["Sum"] = h.GetSampleSum()
		}
		buckets := h.GetBucket()
		barr := make(map[string]generic, len(buckets))
		for _, b := range buckets {
			barr[fmt.Sprintf("=%f", b.GetUpperBound())] = b.GetCumulativeCount()
		}
		val["Buckets"] = barr
		res[idx] = val
	}
	return res
}

func (t *jsonTransformer) exportGauge(metrics []*model.Metric) generic {
	res := make([]generic, len(metrics))
	for idx, m := range metrics {
		val := make(map[string]generic)
		g := m.GetGauge()
		lables := m.GetLabel()
		val["Type"] = "Gauge"
		val["Tags"] = t.exportLables(lables)
		if math.IsNaN(g.GetValue()) {
			val["Value"] = 0.0
		} else {
			val["Value"] = g.GetValue()
		}
		res[idx] = val
	}
	return res
}

func (t *jsonTransformer) exportCounter(metrics []*model.Metric) generic {
	res := make([]generic, len(metrics))
	for idx, m := range metrics {
		val := make(map[string]generic)
		c := m.GetCounter()
		lables := m.GetLabel()
		val["Type"] = "Counter"
		val["Tags"] = t.exportLables(lables)
		if math.IsNaN(c.GetValue()) {
			val["Value"] = 0
		} else {
			val["Value"] = c.GetValue()
		}
		res[idx] = val
	}
	return res
}

func (t *jsonTransformer) quantileKey(val float64) string {
	s := strconv.FormatFloat(val, 'f', 4, 64)
	if s == "0.5000" {
		return "p50"
	} else if s == "0.7500" {
		return "p75"
	} else if s == "0.9500" {
		return "p95"
	} else if s == "0.9900" {
		return "p99"
	} else if s == "0.9990" {
		return "p99.9"
	} else {
		return ""
	}
}

func (t *jsonTransformer) exportSummary(metrics []*model.Metric) generic {
	res := make([]generic, len(metrics))
	for idx, m := range metrics {
		val := make(map[string]generic)
		s := m.GetSummary()
		lables := m.GetLabel()
		val["Type"] = "Summary"
		val["Tags"] = t.exportLables(lables)
		val["Count"] = s.GetSampleCount()
		if math.IsNaN(s.GetSampleSum()) {
			val["Sum"] = 0.0
		} else {
			val["Sum"] = s.GetSampleSum()
		}
		quantiles := s.GetQuantile()
		qarr := make(map[string]float64, len(quantiles))
		for _, quality := range quantiles {
			if math.IsNaN(quality.GetValue()) {
				qarr[t.quantileKey(quality.GetQuantile())] = 0.0
			} else {
				qarr[t.quantileKey(quality.GetQuantile())] = quality.GetValue()
			}
		}
		val["Quantile"] = qarr
		res[idx] = val
	}
	return res
}

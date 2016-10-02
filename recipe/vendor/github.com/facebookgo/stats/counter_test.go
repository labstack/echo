package stats_test

import (
	"testing"

	"github.com/facebookgo/ensure"
	"github.com/facebookgo/stats"
)

func TestSimpleCounterAggregation(t *testing.T) {
	t.Parallel()

	a := stats.Aggregates{}
	a.Add(&stats.SimpleCounter{
		Key:    "foo.avg",
		Values: []float64{1},
		Type:   stats.AggregateAvg,
	})
	a.Add(&stats.SimpleCounter{
		Key:    "foo.sum",
		Values: []float64{1},
		Type:   stats.AggregateSum,
	})
	a.Add(&stats.SimpleCounter{
		Key:    "foo.time",
		Values: []float64{0, 1, 2, 3, 4},
		Type:   stats.AggregateHistogram,
	})
	a.Add(&stats.SimpleCounter{
		Key:    "foo.sum",
		Values: []float64{2},
		Type:   stats.AggregateSum,
	})
	a.Add(&stats.SimpleCounter{
		Key:    "foo.time",
		Values: []float64{5, 6, 7, 8, 9, 10},
		Type:   stats.AggregateHistogram,
	})
	a.Add(&stats.SimpleCounter{
		Key:    "foo.avg",
		Values: []float64{2},
		Type:   stats.AggregateAvg,
	})

	all := map[string]float64{}
	for _, counter := range a {
		for key, value := range counter.(*stats.SimpleCounter).Aggregate() {
			all[key] = value
		}
	}

	ensure.DeepEqual(t, all, map[string]float64{
		"foo.avg":      1.5,
		"foo.sum":      3.0,
		"foo.time":     5.0,
		"foo.time.p50": 5.0,
		"foo.time.p95": 10.0,
		"foo.time.p99": 10.0,
	})
}

package stats_test

import (
	"testing"

	"github.com/facebookgo/ensure"
	"github.com/facebookgo/stats"
)

func TestAverage(t *testing.T) {
	t.Parallel()
	ensure.DeepEqual(t, stats.Average([]float64{}), 0.0)
	ensure.DeepEqual(t, stats.Average([]float64{1}), 1.0)
	ensure.DeepEqual(t, stats.Average([]float64{1, 2}), 1.5)
	ensure.DeepEqual(t, stats.Average([]float64{1, 2, 3}), 2.0)
}

func TestSum(t *testing.T) {
	t.Parallel()
	ensure.DeepEqual(t, stats.Sum([]float64{}), 0.0)
	ensure.DeepEqual(t, stats.Sum([]float64{1}), 1.0)
	ensure.DeepEqual(t, stats.Sum([]float64{1, 2}), 3.0)
	ensure.DeepEqual(t, stats.Sum([]float64{1, 2, 3}), 6.0)
}

func TestPercentiles(t *testing.T) {
	t.Parallel()
	percentiles := map[string]float64{
		"p50": 0.5,
		"p90": 0.9,
	}
	input := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	expected := map[string]float64{
		"p50": 5,
		"p90": 9,
	}
	ensure.DeepEqual(t, stats.Percentiles(input, percentiles), expected)
}

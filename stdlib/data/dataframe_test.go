package data_test

import (
	"testing"

	"AxiomLang/stdlib/data"
)

func TestDataFrame_SumInt64(t *testing.T) {
	df := data.NewDataFrame()

	values := []int64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	df.AddInt64Column("metrics", values)

	sum, err := df.SumInt64("metrics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var expected int64 = 550
	if sum != expected {
		t.Fatalf("expected sum %d, got %d", expected, sum)
	}
}

func TestDataFrame_SumInt64_EdgeCases(t *testing.T) {
	testSizes := []int{0, 1, 2, 7, 8, 9, 15, 16, 17, 63, 64, 65, 1000}
	for _, size := range testSizes {
		df := data.NewDataFrame()
		vals := make([]int64, size)
		var expected int64
		for i := 0; i < size; i++ {
			vals[i] = int64(i + 1)
			expected += vals[i]
		}
		df.AddInt64Column("test", vals)

		sum, err := df.SumInt64("test")
		if err != nil {
			t.Fatalf("unexpected error for size %d: %v", size, err)
		}
		if sum != expected {
			t.Fatalf("size %d: expected sum %d, got %d", size, expected, sum)
		}
	}
}

func BenchmarkDataFrame_SumInt64_AVX2(b *testing.B) {
	df := data.NewDataFrame()
	n := 1000000 // 1 million int64s (8MB)
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = int64(i)
	}
	df.AddInt64Column("large", vals)

	b.ResetTimer()
	b.SetBytes(int64(n * 8))
	for i := 0; i < b.N; i++ {
		_, _ = df.SumInt64("large")
	}
}

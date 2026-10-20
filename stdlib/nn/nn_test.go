package nn_test

import (
	"math"
	"testing"

	"AxiomLang/stdlib/nn"
)

func TestLinearLayer_Forward(t *testing.T) {
	in := 4
	out := 2
	layer := nn.NewLinearLayer(in, out)

	input := nn.NewTensor(in)
	for i := 0; i < in; i++ {
		input.Data[i] = float32(i + 1)
	}

	output := layer.Forward(input)
	if len(output.Data) != out {
		t.Fatalf("expected output dimension %d, got %d", out, len(output.Data))
	}

	for i := 0; i < out; i++ {
		var expected float32
		rowOffset := i * in
		for j := 0; j < in; j++ {
			expected += layer.Weights.Data[rowOffset+j] * input.Data[j]
		}
		expected += layer.Bias.Data[i]

		diff := math.Abs(float64(output.Data[i] - expected))
		if diff > 1e-4 {
			t.Fatalf("row %d: expected %.5f, got %.5f (diff: %e)", i, expected, output.Data[i], diff)
		}
	}
}

func TestLinearLayer_VariousSizes(t *testing.T) {
	sizes := []int{1, 3, 7, 8, 15, 16, 17, 31, 32, 64, 128}
	for _, in := range sizes {
		layer := nn.NewLinearLayer(in, 4)
		input := nn.NewTensor(in)
		for j := 0; j < in; j++ {
			input.Data[j] = 0.5
		}
		output := layer.Forward(input)
		if len(output.Data) != 4 {
			t.Fatalf("failed for size %d", in)
		}
	}
}

func BenchmarkLinearLayer_Forward_AVX_FMA(b *testing.B) {
	in := 512
	out := 256
	layer := nn.NewLinearLayer(in, out)
	input := nn.NewTensor(in)
	for i := 0; i < in; i++ {
		input.Data[i] = 1.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = layer.Forward(input)
	}
}

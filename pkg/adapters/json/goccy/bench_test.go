package goccy_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/armineyvazi/common.git/pkg/adapters/json/goccy"
)

type benchPayload struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Tags     []string `json:"tags"`
	Active   bool     `json:"active"`
	Score    float64  `json:"score"`
	Children []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"children"`
}

var samplePayload = benchPayload{
	ID:     1,
	Name:   "armin",
	Email:  "armin@example.com",
	Tags:   []string{"go", "backend", "infra"},
	Active: true,
	Score:  9.8,
	Children: []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}{
		{1, "child-a"},
		{2, "child-b"},
	},
}

// BenchmarkGoccy_Marshal measures goccy/go-json marshal throughput.
func BenchmarkGoccy_Marshal(b *testing.B) {
	c := goccy.New()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = c.Marshal(samplePayload)
	}
}

// BenchmarkStdlib_Marshal is the baseline for comparison.
func BenchmarkStdlib_Marshal(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = json.Marshal(samplePayload)
	}
}

// BenchmarkGoccy_Unmarshal measures goccy/go-json unmarshal throughput.
func BenchmarkGoccy_Unmarshal(b *testing.B) {
	c := goccy.New()
	data, _ := c.Marshal(samplePayload)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var out benchPayload
		_ = c.Unmarshal(data, &out)
	}
}

// BenchmarkStdlib_Unmarshal is the baseline for comparison.
func BenchmarkStdlib_Unmarshal(b *testing.B) {
	data, _ := json.Marshal(samplePayload)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var out benchPayload
		_ = json.Unmarshal(data, &out)
	}
}

// BenchmarkGoccy_Decoder measures streaming decode throughput.
func BenchmarkGoccy_Decoder(b *testing.B) {
	c := goccy.New()
	data, _ := c.Marshal(samplePayload)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var out benchPayload
		_ = c.NewDecoder(bytes.NewReader(data)).Decode(&out)
	}
}

// BenchmarkGoccy_Marshal_Parallel stresses concurrent marshal paths.
func BenchmarkGoccy_Marshal_Parallel(b *testing.B) {
	c := goccy.New()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.Marshal(samplePayload)
		}
	})
}

// BenchmarkGoccy_Unmarshal_Parallel stresses concurrent unmarshal paths.
func BenchmarkGoccy_Unmarshal_Parallel(b *testing.B) {
	c := goccy.New()
	data, _ := c.Marshal(samplePayload)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var out benchPayload
			_ = c.Unmarshal(data, &out)
		}
	})
}

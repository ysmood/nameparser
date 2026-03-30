package nameparser_test

import (
	"testing"

	"github.com/ysmood/nameparser"
)

var benchmarkInputs = []string{
	"John Doe",
	"Doe, Dr. John A. Kenneth, Jr.",
	"Rev John A. Kenneth Doe III (Kenny)",
	"de la Cruz, Maria",
	"dr.mr John Doe",
	"John and of Doe",
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = nameparser.New(benchmarkInputs[i%len(benchmarkInputs)])
	}
}

func BenchmarkNewWithSharedCustomConstants(b *testing.B) {
	b.ReportAllocs()
	c := nameparser.NewConstants()
	for i := 0; i < b.N; i++ {
		_ = nameparser.NewWithConstants(benchmarkInputs[i%len(benchmarkInputs)], c)
	}
}

func BenchmarkNewWithFreshConstants(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := nameparser.NewConstants()
		_ = nameparser.NewWithConstants(benchmarkInputs[i%len(benchmarkInputs)], c)
	}
}

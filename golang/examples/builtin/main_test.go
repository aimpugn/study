package main

import "testing"

func BenchmarkRemoveHyphensLoop(b *testing.B) {
	uuid := "123e4567-e89b-12d3-a456-426614174000"
	for i := 0; i < b.N; i++ {
		removeHyphensLoop(uuid)
	}
	b.ReportAllocs()
}

func BenchmarkRemoveHyphensReplaceAll(b *testing.B) {
	uuid := "123e4567-e89b-12d3-a456-426614174000"
	for i := 0; i < b.N; i++ {
		removeHyphensReplaceAll(uuid)
	}
	b.ReportAllocs()
}

package logger

import (
	"fmt"
)

type Stats struct {
	resolveCount       int
	totalResolveCount  int
	downloadCount      int
	totalDownloadCount int
}

func (s *Stats) PrettyPrintStats() {
	// TODO: figure out a way to implement proper printing without inputting all the four numbers all the time
	fmt.Printf("\r🔍[%d/%d] 🚚[%d/%d]", s.resolveCount, s.totalResolveCount, s.downloadCount, s.totalDownloadCount)
}

func (s *Stats) IncrementResolveCount() {
	s.resolveCount += 1
	s.PrettyPrintStats()
}

func (s *Stats) IncrementTotalResolveCount() {
	s.totalResolveCount += 1
	s.PrettyPrintStats()
}

func (s *Stats) IncrementDownloadCount() {
	s.downloadCount += 1
	s.PrettyPrintStats()
}

func (s *Stats) IncrementTotalDownloadCount() {
	s.totalDownloadCount += 1
	s.PrettyPrintStats()
}

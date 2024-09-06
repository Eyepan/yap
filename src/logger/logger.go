package logger

import (
	"fmt"
	"sync"
)

type Stats struct {
	ResolveCount       int
	TotalResolveCount  int
	DownloadCount      int
	TotalDownloadCount int
	statsMu            sync.Mutex
}

func (s *Stats) PrettyPrintStats() {
	// TODO: figure out a way to implement proper printing without inputting all the four numbers all the time
	fmt.Printf("\r🔍[%d/%d] 🚚[%d/%d]\t", s.ResolveCount, s.TotalResolveCount, s.DownloadCount, s.TotalDownloadCount)
}

func (s *Stats) IncrementResolveCount() {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.ResolveCount += 1
	s.PrettyPrintStats()
}

func (s *Stats) IncrementTotalResolveCount() {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.TotalResolveCount += 1
	s.PrettyPrintStats()
}

func (s *Stats) IncrementDownloadCount() {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.DownloadCount += 1
	s.PrettyPrintStats()
}

func (s *Stats) IncrementTotalDownloadCount() {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.TotalDownloadCount += 1
	s.PrettyPrintStats()
}

func PrintCurrentCommand(command string) {
	fmt.Println("\033[1m yap " + command + "\033[0m")
}

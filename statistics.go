package main

import (
	"fmt"
	"sync"
	"time"
)

type Statistics struct {
	duplicatedCount, totalCount int64

	blockTimestamps map[uint64]int64
	txTimestamps    map[uint64][]int64

	processorLagSum   int64
	processorLagMax   int64
	processorLagCount int64

	mu sync.Mutex
}

func newStatistics() *Statistics {
	return &Statistics{
		blockTimestamps: make(map[uint64]int64),
		txTimestamps:    make(map[uint64][]int64),
	}
}

func (s *Statistics) add(blockTs int64, slot uint64, duplicated bool, timestamp time.Time, processingTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if duplicated {
		s.duplicatedCount++
	}
	s.totalCount++

	if blockTs > 0 {
		s.blockTimestamps[slot] = blockTs * 1000
	}

	if !duplicated {
		s.txTimestamps[slot] = append(s.txTimestamps[slot], timestamp.UnixMilli())
	}

}

func (s *Statistics) report() {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Printf("-----------------------------------------------------------\n")
	fmt.Printf("total txs processed: %d duplicate transactions: %d (%.1f %%) \n", s.totalCount, s.duplicatedCount,
		float64(s.duplicatedCount)*100/float64(s.totalCount))

	count := int64(0)
	sumLag := int64(0)
	maxLag := int64(0)
	for slot, txts := range s.txTimestamps {
		slotTs, hasTx := s.blockTimestamps[slot]
		if !hasTx {
			continue
		}
		for _, ts := range txts {
			lag := ts - slotTs
			if lag > maxLag {
				maxLag = lag
			}
			sumLag += lag
			count++
		}
	}

	if count > 0 {
		fmt.Printf("Average lag to block time %d msec, max lag %d msec\n", sumLag/count, maxLag)
	}

	if s.processorLagCount > 0 {
		fmt.Printf("Average lag to message time %d msec, max lag %d msec\n", s.processorLagSum/s.processorLagCount, s.processorLagMax)
	}
	fmt.Printf("-----------------------------------------------------------\n")
}

func (s *Statistics) record(timestamp time.Time, processingTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lag := processingTime.Sub(timestamp).Milliseconds()

	s.processorLagCount++
	s.processorLagSum += lag

	if lag > s.processorLagMax {
		s.processorLagMax = lag
	}

}

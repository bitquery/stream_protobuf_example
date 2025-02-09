package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/errgroup"
)

type ProcessorConfig struct {
	Buffer  int
	Workers int
}

type Listener interface {
	enqueue(message *kafka.Message)
}

type dedupCache struct {
	cache *expirable.LRU[string, bool]
	mu    sync.Mutex
}

type processFn func(context.Context, *kafka.Message, int, *dedupCache) error

type Processor struct {
	queue     chan (*kafka.Message)
	wg        errgroup.Group
	config    ProcessorConfig
	processFn processFn
	stat      *Statistics
	dedup     *dedupCache
}

func newProcessor(config *Config) (*Processor, error) {
	processor := &Processor{
		queue:  make(chan *kafka.Message, config.Processor.Buffer),
		config: config.Processor,
		stat:   newStatistics(),
		dedup: &dedupCache{
			cache: expirable.NewLRU[string, bool](10000, nil,
				time.Second*time.Duration(240)),
		},
	}

	var processFn processFn
	switch config.Consumer.Topic {
	case "solana.dextrades.proto":
		processFn = processor.dexTradesMessageHandler
	case "solana.transactions.proto":
		processFn = processor.transactionsMessageHandler
	case "solana.tokens.proto":
		processFn = processor.tokensMessageHandler
	default:
		processFn = processor.jsonMessageHandler
	}

	processor.processFn = processFn
	return processor, nil
}

func (processor *Processor) enqueue(message *kafka.Message) {
	processor.queue <- message
}

func (processor *Processor) start(ctx context.Context) {
	counter := 0
	for i := 0; i < processor.config.Workers; i++ {
		processor.wg.Go(func() error {
			i := i
			fmt.Println("Starting worker ", i)
			for {
				select {
				case <-ctx.Done():
					fmt.Println("Done, exiting processor loop worker ", i)
					return nil
				case message := <-processor.queue:
					err := processor.processFn(ctx, message, i, processor.dedup)
					if err != nil {
						fmt.Println("Error processing message", err)
					}
					counter++
					if counter%100 == 0 {
						processor.stat.report()
					}
				}
			}
			return nil
		})
	}
}

func (processor *Processor) close() {
	fmt.Println("Shutting down processor...")
	processor.wg.Wait()
	fmt.Println("Processor stopped")
	processor.stat.report()
}

func (dedup *dedupCache) isDuplicated(slot uint64, index uint32) bool {
	key := fmt.Sprintf("%d-%d", slot, index)
	dedup.mu.Lock()
	defer dedup.mu.Unlock()

	if dedup.cache.Contains(key) {
		return true
	}

	dedup.cache.Add(key, true)
	return false
}

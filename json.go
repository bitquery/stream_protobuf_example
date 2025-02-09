package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type JsonMessage struct {
	Block struct {
		Hash      string `json:"Hash"`
		Height    uint64 `json:"Height"`
		Slot      uint64 `json:"Slot"`
		Timestamp int64  `json:"Timestamp"`
	} `json:"Block"`
	Transaction struct {
		Index uint32 `json:"Index"`
	} `json:"Transaction"`
}

func (processor *Processor) jsonMessageHandler(_ context.Context,
	message *kafka.Message, worker int, dedup *dedupCache) error {
	processingTime := time.Now()
	processor.stat.record(message.Timestamp, processingTime)

	var batch []JsonMessage

	err := json.Unmarshal(message.Value, &batch)
	if err != nil {
		return err
	}

	duplicated := 0
	for _, t := range batch {
		isDuplicated := dedup.isDuplicated(t.Block.Slot, t.Transaction.Index)
		processor.stat.add(t.Block.Timestamp, t.Block.Slot, isDuplicated, message.Timestamp, processingTime)
		if isDuplicated {
			duplicated += 1
		}
	}

	fmt.Printf("slot %d processed with lag %d msec %d txs, %d duplicated from partition %d[%s] in worker %d\n",
		batch[0].Block.Slot, processingTime.Sub(message.Timestamp).Milliseconds(),
		len(batch), duplicated,
		message.TopicPartition.Partition, message.TopicPartition.Offset, worker)

	return nil
}

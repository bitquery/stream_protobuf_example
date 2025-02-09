package main

import (
	"context"
	"fmt"
	"time"

	solana_messages "github.com/bitquery/streaming_protobuf/v2/solana/messages"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/protobuf/proto"
)

func (processor *Processor) dexTradesMessageHandler(_ context.Context,
	message *kafka.Message, worker int, dedup *dedupCache) error {
	processingTime := time.Now()
	processor.stat.record(message.Timestamp, processingTime)

	var batch solana_messages.DexParsedBlockMessage
	err := proto.Unmarshal(message.Value, &batch)
	if err != nil {
		return err
	}

	count := 0
	duplicated := 0
	for _, t := range batch.Transactions {
		isDuplicated := dedup.isDuplicated(batch.Header.Slot, t.Index)
		processor.stat.add(batch.Header.Timestamp, batch.Header.Slot, isDuplicated, message.Timestamp, processingTime)
		count += len(t.Trades)
		if isDuplicated {
			duplicated += 1
		}
	}

	fmt.Printf("slot %d processed with lag %d msec %d txs, %d duplicated (%d trades) from partition %d[%s] in worker %d\n",
		batch.Header.Slot, processingTime.Sub(message.Timestamp).Milliseconds(),
		len(batch.Transactions), duplicated, count,
		message.TopicPartition.Partition, message.TopicPartition.Offset, worker)

	return nil
}

func (processor *Processor) transactionsMessageHandler(_ context.Context,
	message *kafka.Message, worker int, dedup *dedupCache) error {
	processingTime := time.Now()
	processor.stat.record(message.Timestamp, processingTime)

	var batch solana_messages.ParsedIdlBlockMessage
	err := proto.Unmarshal(message.Value, &batch)
	if err != nil {
		return err
	}
	duplicated := 0
	for _, t := range batch.Transactions {
		isDuplicated := dedup.isDuplicated(batch.Header.Slot, t.Index)
		processor.stat.add(batch.Header.Timestamp, batch.Header.Slot, isDuplicated, message.Timestamp, processingTime)
		if isDuplicated {
			duplicated += 1
		}
	}
	fmt.Printf("slot %d processed with lag %d msec %d txs, %d duplicated from partition %d[%s] in worker %d\n",
		batch.Header.Slot, processingTime.Sub(message.Timestamp).Milliseconds(),
		len(batch.Transactions), duplicated,
		message.TopicPartition.Partition, message.TopicPartition.Offset, worker)
	return nil
}

func (processor *Processor) tokensMessageHandler(_ context.Context,
	message *kafka.Message, worker int, dedup *dedupCache) error {
	processingTime := time.Now()
	processor.stat.record(message.Timestamp, processingTime)

	var batch solana_messages.TokenBlockMessage
	err := proto.Unmarshal(message.Value, &batch)
	if err != nil {
		return err
	}

	transfers := 0
	balanceUpdates := 0
	instructionBalanceUpdates := 0
	instructionTokenSupplyUpdates := 0

	duplicated := 0
	for _, t := range batch.Transactions {
		isDuplicated := dedup.isDuplicated(batch.Header.Slot, t.Index)
		processor.stat.add(batch.Header.Timestamp, batch.Header.Slot, isDuplicated, message.Timestamp, processingTime)
		transfers += len(t.Transfers)
		balanceUpdates += len(t.BalanceUpdates)
		for _, instr := range t.InstructionBalanceUpdates {
			instructionBalanceUpdates += len(instr.OwnCurrencyBalanceUpdates)
			instructionTokenSupplyUpdates += len(instr.TokenSupplyUpdates)
		}
		if isDuplicated {
			duplicated += 1
		}
	}

	fmt.Printf("slot %d processed with lag %d msec %d txs, %d duplicated (%d transfers %d balanceUpdates %d instructionBalanceUpdates %d instructionTokenSupplyUpdates) from partition %d[%s] in worker %d\n",
		batch.Header.Slot, processingTime.Sub(message.Timestamp).Milliseconds(),
		len(batch.Transactions), duplicated,
		transfers, balanceUpdates, instructionBalanceUpdates, instructionTokenSupplyUpdates,
		message.TopicPartition.Partition, message.TopicPartition.Offset, worker)

	return nil
}

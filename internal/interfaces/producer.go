package interfaces

import (
	"context"
)

// DataProducer определяет контракт для отправки данных во внешние системы (например, Kafka)
type DataProducer interface {
	Produce(ctx context.Context, key, value []byte) error
	Close() error
}

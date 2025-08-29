package interfaces

import (
	"context"
)

// KafkaService определяет контракт для отправки данных во внешние системы
type KafkaService interface {
	Produce(ctx context.Context, key, value []byte) error
	Close() error
}

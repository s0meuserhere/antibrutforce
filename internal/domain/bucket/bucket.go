package bucket

import (
	"errors"
	"fmt"
	"time"

	"github.com/s0meuserhere/antibrutforce/internal/domain/limits"
)

// ErrInvalidCapacity означает ёмкость вне диапазона от 1 до limits.Max.
var ErrInvalidCapacity = errors.New("bucket capacity is out of range")

// Bucket - leaky bucket попыток. Ёмкость равна лимиту в минуту,
// за минуту вытекает вся ёмкость, освобождается только целая попытка.
type Bucket struct {
	capacity int
	level    int
	leakedAt time.Time
}

// New возвращает пустой бакет с ёмкостью от 1 до limits.Max.
func New(capacity int) (*Bucket, error) {
	if capacity < 1 || capacity > limits.Max {
		return nil, fmt.Errorf("capacity %d: %w", capacity, ErrInvalidCapacity)
	}

	return &Bucket{capacity: capacity}, nil
}

// Restore собирает бакет из сохранённого состояния без проверок New.
func Restore(capacity, level int, leakedAt time.Time) *Bucket {
	return &Bucket{capacity: capacity, level: level, leakedAt: leakedAt}
}

// Capacity возвращает ёмкость бакета.
func (b *Bucket) Capacity() int {
	return b.capacity
}

// Level возвращает число попыток в бакете.
func (b *Bucket) Level() int {
	return b.level
}

// LeakedAt возвращает момент, от которого отсчитывается утечка.
func (b *Bucket) LeakedAt() time.Time {
	return b.leakedAt
}

// Allow выпускает попытки, утёкшие к моменту now, и принимает новую, если есть место.
// Вытекают только целые попытки: остаток времени с leakedAt переходит на следующую.
func (b *Bucket) Allow(now time.Time) bool {
	elapsed := now.Sub(b.leakedAt)
	leaked := int(int64(elapsed) * int64(b.capacity) / int64(time.Minute))

	switch {
	case b.level == 0 || elapsed >= b.drain(b.level):
		// Бакет опустел: отсчёт утечки начинается заново с now.
		b.level = 0
		b.leakedAt = now
	case leaked > 0:
		// Вытекли целые попытки: leakedAt сдвигается ровно на время их утечки.
		b.level -= leaked
		b.leakedAt = b.leakedAt.Add(b.drain(leaked))
	}

	if b.level >= b.capacity {
		return false
	}

	b.level++

	return true
}

// KeepFor возвращает, через сколько после now бакет опустеет.
func (b *Bucket) KeepFor(now time.Time) time.Duration {
	emptyAt := b.leakedAt.Add(b.drain(b.level))
	if !emptyAt.After(now) {
		return 0
	}

	return emptyAt.Sub(now)
}

// drain возвращает время, за которое из бакета вытекает n попыток, с округлением вверх.
func (b *Bucket) drain(n int) time.Duration {
	total := int64(n) * int64(time.Minute)
	capacity := int64(b.capacity)

	return time.Duration((total + capacity - 1) / capacity)
}

package bucket

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/s0meuserhere/antibrutforce/internal/domain/attempt"
	"github.com/s0meuserhere/antibrutforce/internal/domain/bucket"
	"github.com/s0meuserhere/antibrutforce/internal/domain/limits"
	"github.com/s0meuserhere/antibrutforce/internal/domain/reset"
)

// Repository хранит бакеты логинов, паролей и IP в Redis.
type Repository struct {
	client *redis.Client
}

// rBucket - бакет в Redis. Ёмкость не хранится: она приходит из лимитов.
type rBucket struct {
	Level    int   `json:"level"`
	LeakedAt int64 `json:"leakedAt"`
}

// New возвращает репозиторий бакетов.
func New(client *redis.Client) *Repository {
	return &Repository{client: client}
}

// Update читает бакеты логина, пароля и IP попытки, передаёт их в fn,
// сохраняет результат с TTL до опустения и возвращает ответ fn. Чтение и запись
// атомарны и между экземплярами сервиса: если бакеты изменил параллельный
// запрос, всё повторяется, и fn вызывается снова. Повторы идут, пока не истечёт
// ctx, поэтому вызывающий задаёт ему срок.
func (r *Repository) Update(
	ctx context.Context,
	a attempt.Attempt,
	l limits.Limits,
	now time.Time,
	fn func(buckets attempt.Buckets) bool,
) (bool, error) {
	login, password, ip := bucketKeys(a)

	for {
		// Watch возвращает только ошибку, поэтому ответ fn выносится через замыкание.
		var ok bool

		err := r.client.Watch(ctx, func(tx *redis.Tx) error {
			result, err := update(ctx, tx, a, l, now, fn)
			ok = result

			return err
		}, login, password, ip)
		if errors.Is(err, redis.TxFailedErr) {
			continue
		}

		if err != nil {
			return false, fmt.Errorf("update buckets: %w", err)
		}

		return ok, nil
	}
}

// Reset удаляет бакеты логина и IP цели.
func (r *Repository) Reset(ctx context.Context, t reset.Target) error {
	if err := r.client.Del(ctx, loginKey(t.Login()), ipKey(t.IP())).Err(); err != nil {
		return fmt.Errorf("delete buckets: %w", err)
	}

	return nil
}

func update(
	ctx context.Context,
	tx *redis.Tx,
	a attempt.Attempt,
	l limits.Limits,
	now time.Time,
	fn func(buckets attempt.Buckets) bool,
) (bool, error) {
	login, password, ip := bucketKeys(a)

	values, err := tx.MGet(ctx, login, password, ip).Result()
	if err != nil {
		return false, fmt.Errorf("read buckets: %w", err)
	}

	// Значения MGET идут в порядке ключей: логин, пароль, IP.
	loginBucket, err := load(values[0], l.PerLogin())
	if err != nil {
		return false, err
	}

	passwordBucket, err := load(values[1], l.PerPassword())
	if err != nil {
		return false, err
	}

	ipBucket, err := load(values[2], l.PerIP())
	if err != nil {
		return false, err
	}

	ok := fn(attempt.NewBuckets(loginBucket, passwordBucket, ipBucket))

	_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if err := save(ctx, pipe, login, loginBucket, now); err != nil {
			return err
		}

		if err := save(ctx, pipe, password, passwordBucket, now); err != nil {
			return err
		}

		return save(ctx, pipe, ip, ipBucket, now)
	})
	if err != nil {
		return false, fmt.Errorf("save buckets: %w", err)
	}

	return ok, nil
}

// load собирает бакет из значения в Redis. Отсутствующий ключ даёт пустой бакет.
func load(value any, capacity int) (*bucket.Bucket, error) {
	data, ok := value.(string)
	if !ok {
		b, err := bucket.New(capacity)
		if err != nil {
			return nil, fmt.Errorf("new bucket: %w", err)
		}

		return b, nil
	}

	var rb rBucket
	if err := json.Unmarshal([]byte(data), &rb); err != nil {
		return nil, fmt.Errorf("decode bucket: %w", err)
	}

	return bucket.Restore(capacity, rb.Level, time.Unix(0, rb.LeakedAt)), nil
}

// save ставит запись бакета в транзакцию. Ошибки команд вернёт EXEC.
// Ключ живёт, пока бакет не опустеет. TTL округляется вверх до миллисекунды:
// Redis не хранит его точнее, а ключ не должен исчезнуть раньше времени.
func save(ctx context.Context, pipe redis.Pipeliner, key string, b *bucket.Bucket, now time.Time) error {
	ttl := (b.KeepFor(now) + time.Millisecond - 1).Truncate(time.Millisecond)
	if ttl == 0 {
		pipe.Del(ctx, key)

		return nil
	}

	data, err := json.Marshal(rBucket{Level: b.Level(), LeakedAt: b.LeakedAt().UnixNano()})
	if err != nil {
		return fmt.Errorf("encode bucket: %w", err)
	}

	pipe.Set(ctx, key, data, ttl)

	return nil
}

// bucketKeys возвращает ключи бакетов попытки в Redis.
func bucketKeys(a attempt.Attempt) (login, password, ip string) {
	return loginKey(a.Login()), passwordKey(a.Password()), ipKey(a.IP())
}

func loginKey(login string) string {
	return "bucket:login:" + login
}

// passwordKey хранит в ключе SHA-256 пароля, сам пароль в Redis не попадает.
func passwordKey(password string) string {
	sum := sha256.Sum256([]byte(password))

	return "bucket:password:" + hex.EncodeToString(sum[:])
}

func ipKey(ip netip.Addr) string {
	return "bucket:ip:" + ip.String()
}

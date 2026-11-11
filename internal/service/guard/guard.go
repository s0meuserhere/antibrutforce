package guard

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/s0meuserhere/antibrutforce/internal/domain/attempt"
	"github.com/s0meuserhere/antibrutforce/internal/domain/limits"
	"github.com/s0meuserhere/antibrutforce/internal/domain/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/domain/reset"
)

// ErrInvalidArgument - запрос не прошёл проверку. Остальные ошибки сервиса - сбои хранилищ.
var ErrInvalidArgument = errors.New("invalid argument")

// Lists хранит whitelist и blacklist.
type Lists interface {
	// Get возвращает whitelist и blacklist вместе.
	Get(ctx context.Context) (netlist.Lists, error)
	Add(ctx context.Context, e netlist.Entry) error
	Remove(ctx context.Context, e netlist.Entry) error
}

// Buckets хранит бакеты логинов, паролей и IP.
type Buckets interface {
	// Update передаёт в fn бакеты попытки, сохраняет то,
	// что fn с ними сделала, и возвращает ответ fn.
	//
	// Чтение и сохранение должны быть атомарны для всех экземпляров сервиса,
	// а не только внутри одного процесса: иначе параллельные запросы с разных
	// экземпляров прочитают одно и то же состояние и пропустят попыток больше
	// лимита. Если бакеты изменил параллельный запрос, fn вызывается снова
	// со свежими бакетами.
	Update(
		ctx context.Context,
		a attempt.Attempt,
		l limits.Limits,
		now time.Time,
		fn func(buckets attempt.Buckets) bool,
	) (bool, error)
	// Reset удаляет бакеты цели сброса.
	Reset(ctx context.Context, t reset.Target) error
}

// Service - сценарии анти-брутфорса.
type Service struct {
	lists   Lists
	buckets Buckets
	limits  limits.Limits
	now     func() time.Time
}

// New возвращает сервис. now - источник текущего времени.
func New(lists Lists, buckets Buckets, l limits.Limits, now func() time.Time) *Service {
	return &Service{lists: lists, buckets: buckets, limits: l, now: now}
}

// Check решает, пропустить ли попытку авторизации. Адрес из whitelist
// пропускается, из blacklist - нет, в обоих случаях бакеты не трогаются.
// Остальные попытки учитываются в бакетах логина, пароля и IP.
func (s *Service) Check(ctx context.Context, login, password, ip string) (bool, error) {
	a, err := attempt.New(login, password, ip)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	lists, err := s.lists.Get(ctx)
	if err != nil {
		return false, fmt.Errorf("read lists: %w", err)
	}

	if ok, listed := a.AllowByLists(lists); listed {
		return ok, nil
	}

	now := s.now()

	ok, err := s.buckets.Update(ctx, a, s.limits, now, func(buckets attempt.Buckets) bool {
		return a.AllowByBuckets(now, buckets)
	})
	if err != nil {
		return false, fmt.Errorf("update buckets: %w", err)
	}

	return ok, nil
}

// Reset сбрасывает бакеты логина и IP. Нужно указать хотя бы одно из них.
func (s *Service) Reset(ctx context.Context, login, ip string) error {
	t, err := reset.New(login, ip)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	if err := s.buckets.Reset(ctx, t); err != nil {
		return fmt.Errorf("reset buckets: %w", err)
	}

	return nil
}

// Add добавляет подсеть в список. Адрес хоста в подсети обнуляется:
// 192.1.1.10/25 сохраняется как 192.1.1.0/25.
func (s *Service) Add(ctx context.Context, kind netlist.Kind, cidr string) error {
	e, err := netlist.NewEntry(kind, cidr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	if err := s.lists.Add(ctx, e); err != nil {
		return fmt.Errorf("add subnet: %w", err)
	}

	return nil
}

// Remove удаляет подсеть из списка. Подсеть приводится к виду, в котором её сохранил Add.
func (s *Service) Remove(ctx context.Context, kind netlist.Kind, cidr string) error {
	e, err := netlist.NewEntry(kind, cidr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	if err := s.lists.Remove(ctx, e); err != nil {
		return fmt.Errorf("remove subnet: %w", err)
	}

	return nil
}

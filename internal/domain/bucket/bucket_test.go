package bucket_test

import (
	"errors"
	"testing"
	"time"

	"github.com/s0meuserhere/antibrutforce/internal/domain/bucket"
	"github.com/s0meuserhere/antibrutforce/internal/domain/limits"
)

var t0 = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

type step struct {
	at   time.Duration
	want bool
}

func repeat(n int, at time.Duration, want bool) []step {
	steps := make([]step, n)
	for i := range steps {
		steps[i] = step{at: at, want: want}
	}

	return steps
}

func join(parts ...[]step) []step {
	var steps []step
	for _, part := range parts {
		steps = append(steps, part...)
	}

	return steps
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		capacity int
		wantErr  error
	}{
		{name: "lower bound", capacity: 1},
		{name: "default login limit", capacity: 10},
		{name: "upper bound", capacity: limits.Max},
		{name: "zero", capacity: 0, wantErr: bucket.ErrInvalidCapacity},
		{name: "negative", capacity: -1, wantErr: bucket.ErrInvalidCapacity},
		{name: "above max", capacity: limits.Max + 1, wantErr: bucket.ErrInvalidCapacity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := bucket.New(tt.capacity)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if got.Capacity() != tt.capacity || got.Level() != 0 {
				t.Errorf("New() capacity/level = %d/%d, want %d/0", got.Capacity(), got.Level(), tt.capacity)
			}
		})
	}
}

func TestAllow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		capacity int
		steps    []step
	}{
		{
			name:     "accepts up to capacity",
			capacity: 3,
			steps:    join(repeat(3, 0, true), repeat(1, 0, false)),
		},
		{
			name:     "frees one attempt every minute divided by capacity",
			capacity: 10,
			steps: join(repeat(10, 0, true), []step{
				{at: 5999 * time.Millisecond, want: false},
				{at: 6 * time.Second, want: true},
				{at: 6 * time.Second, want: false},
				{at: 12 * time.Second, want: true},
			}),
		},
		{
			name:     "frees only whole attempt",
			capacity: 7,
			steps: join(repeat(7, 0, true), []step{
				{at: 8571428571 * time.Nanosecond, want: false},
				{at: 8571428572 * time.Nanosecond, want: true},
			}),
		},
		{
			name:     "keeps partial leak between calls",
			capacity: 10,
			steps: join(repeat(10, 0, true), []step{
				{at: 9 * time.Second, want: true},
				{at: 9 * time.Second, want: false},
				{at: 12 * time.Second, want: true},
			}),
		},
		{
			name:     "drains completely in a minute",
			capacity: 10,
			steps:    join(repeat(10, 0, true), repeat(10, time.Minute, true), repeat(1, time.Minute, false)),
		},
		{
			name:     "rejected attempt takes no place",
			capacity: 2,
			steps: join(repeat(2, 0, true), repeat(3, 0, false), []step{
				{at: 30 * time.Second, want: true},
				{at: 30 * time.Second, want: false},
			}),
		},
		{
			name:     "clock going back leaks nothing",
			capacity: 2,
			steps: join(repeat(2, 10*time.Second, true), []step{
				{at: 5 * time.Second, want: false},
				{at: 10 * time.Second, want: false},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b, err := bucket.New(tt.capacity)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			for i, s := range tt.steps {
				if got := b.Allow(t0.Add(s.at)); got != s.want {
					t.Fatalf("step %d: Allow(t0+%v) = %v, want %v", i, s.at, got, s.want)
				}
			}
		})
	}
}

func TestRestore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		capacity int
		level    int
		steps    []step
	}{
		{
			name:     "full bucket",
			capacity: 10,
			level:    10,
			steps: []step{
				{at: 0, want: false},
				{at: 6 * time.Second, want: true},
			},
		},
		{
			name:     "level above lowered capacity",
			capacity: 5,
			level:    8,
			steps: []step{
				{at: 36 * time.Second, want: false},
				{at: 48 * time.Second, want: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := bucket.Restore(tt.capacity, tt.level, t0)
			if b.Capacity() != tt.capacity || b.Level() != tt.level || !b.LeakedAt().Equal(t0) {
				t.Fatalf("Restore() = %d/%d/%v, want %d/%d/%v",
					b.Capacity(), b.Level(), b.LeakedAt(), tt.capacity, tt.level, t0)
			}

			for i, s := range tt.steps {
				if got := b.Allow(t0.Add(s.at)); got != s.want {
					t.Fatalf("step %d: Allow(t0+%v) = %v, want %v", i, s.at, got, s.want)
				}
			}
		})
	}
}

func TestKeepFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		capacity int
		attempts int
		now      time.Duration
		want     time.Duration
	}{
		{name: "empty", capacity: 10, attempts: 0, now: 0, want: 0},
		{name: "one attempt", capacity: 10, attempts: 1, now: 0, want: 6 * time.Second},
		{name: "full", capacity: 10, attempts: 10, now: 0, want: time.Minute},
		{name: "full later", capacity: 10, attempts: 10, now: 15 * time.Second, want: 45 * time.Second},
		{name: "already drained", capacity: 10, attempts: 1, now: 7 * time.Second, want: 0},
		{name: "rounds up", capacity: 7, attempts: 1, now: 0, want: 8571428572 * time.Nanosecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b, err := bucket.New(tt.capacity)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			for range tt.attempts {
				b.Allow(t0)
			}

			if got := b.KeepFor(t0.Add(tt.now)); got != tt.want {
				t.Errorf("KeepFor(t0+%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

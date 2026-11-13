//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	antibfv1 "github.com/s0meuserhere/antibrutforce/gen/antibf/v1"
)

// readyTimeout - сколько ждать, пока сервис подключится к хранилищам и начнёт отвечать.
const readyTimeout = 30 * time.Second

var (
	client antibfv1.AntiBruteforceServiceClient

	// Лимиты, с которыми запущен сервис.
	limitLogin, limitPassword, limitIP int

	// seq делает логины, пароли и адреса разными во всех тестах.
	seq atomic.Uint32
)

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: %v\n", err)
		os.Exit(1)
	}

	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	var err error

	if limitLogin, err = strconv.Atoi(os.Getenv("LIMIT_LOGIN")); err != nil {
		return 0, fmt.Errorf("LIMIT_LOGIN: %w", err)
	}

	if limitPassword, err = strconv.Atoi(os.Getenv("LIMIT_PASSWORD")); err != nil {
		return 0, fmt.Errorf("LIMIT_PASSWORD: %w", err)
	}

	if limitIP, err = strconv.Atoi(os.Getenv("LIMIT_IP")); err != nil {
		return 0, fmt.Errorf("LIMIT_IP: %w", err)
	}

	conn, err := grpc.NewClient(os.Getenv("ANTIBF_GRPC_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return 0, fmt.Errorf("grpc client: %w", err)
	}
	defer conn.Close()

	client = antibfv1.NewAntiBruteforceServiceClient(conn)

	if err := waitReady(); err != nil {
		return 0, err
	}

	return m.Run(), nil
}

// waitReady ждёт, пока сервис ответит. Compose запускает тесты сразу после старта
// контейнера сервиса, а слушать порт он начинает только после подключения к хранилищам.
func waitReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), readyTimeout)
	defer cancel()

	for {
		_, err := client.ResetV1(ctx, &antibfv1.ResetV1Request{Login: "readiness-probe"})
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("service is not ready: %w", err)
		case <-time.After(200 * time.Millisecond):
		}
	}
}

// unique возвращает строку, которой ещё не было в тестах: бакеты разных тестов не пересекаются.
func unique(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, seq.Add(1))
}

// uniqueIP возвращает адрес из 198.18.0.0/15, которого ещё не было в тестах.
// Подсети для списков тесты берут из других диапазонов.
func uniqueIP() string {
	n := seq.Add(1)

	return fmt.Sprintf("198.%d.%d.%d", 18+n>>16&1, n>>8&0xff, n&0xff)
}

// attempt - попытка авторизации.
type attempt struct {
	login    string
	password string
	ip       string
}

// newAttempt возвращает попытку, у которой логин, пароль и адрес ещё не встречались.
func newAttempt() attempt {
	return attempt{login: unique("login"), password: unique("password"), ip: uniqueIP()}
}

// check проверяет попытку и возвращает ответ сервиса.
func check(t *testing.T, a attempt) bool {
	t.Helper()

	resp, err := client.CheckV1(t.Context(), &antibfv1.CheckV1Request{Login: a.login, Password: a.password, Ip: a.ip})
	if err != nil {
		t.Fatalf("check %s %s: %v", a.login, a.ip, err)
	}

	return resp.GetOk()
}

// exhaust делает limit попыток и проверяет, что следующая упирается в лимит.
// next возвращает очередную попытку.
func exhaust(t *testing.T, limit int, next func() attempt) {
	t.Helper()

	for i := range limit {
		if !check(t, next()) {
			t.Fatalf("attempt %d of %d is denied", i+1, limit)
		}
	}

	if check(t, next()) {
		t.Fatalf("attempt %d over limit %d is allowed", limit+1, limit)
	}
}

// rejected проверяет, что запрос отклонён с кодом InvalidArgument,
// и возвращает reason из google.rpc.ErrorInfo в деталях статуса.
func rejected(t *testing.T, err error) string {
	t.Helper()

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("want InvalidArgument, got %v", err)

		return ""
	}

	for _, detail := range st.Details() {
		if info, ok := detail.(*errdetails.ErrorInfo); ok {
			return info.GetReason()
		}
	}

	t.Fatalf("no ErrorInfo in status details of %q", st.Message())

	return ""
}

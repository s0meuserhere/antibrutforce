package grpcserver

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	antibfv1 "github.com/s0meuserhere/antibrutforce/gen/antibf/v1"
	"github.com/s0meuserhere/antibrutforce/internal/domain/attempt"
	"github.com/s0meuserhere/antibrutforce/internal/domain/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/domain/reset"
	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
	"github.com/s0meuserhere/antibrutforce/internal/service/guard"
)

type handler struct {
	antibfv1.UnimplementedAntiBruteforceServiceServer

	service *guard.Service
	logger  *slog.Logger
}

func (h *handler) CheckV1(ctx context.Context, req *antibfv1.CheckV1Request) (*antibfv1.CheckV1Response, error) {
	ok, err := h.service.Check(ctx, req.GetLogin(), req.GetPassword(), req.GetIp())
	if err == nil {
		return &antibfv1.CheckV1Response{Ok: ok}, nil
	}

	reason := antibfv1.CheckV1Response_UNKNOWN

	switch {
	case errors.Is(err, attempt.ErrEmptyLogin):
		reason = antibfv1.CheckV1Response_EMPTY_LOGIN
	case errors.Is(err, attempt.ErrEmptyPassword):
		reason = antibfv1.CheckV1Response_EMPTY_PASSWORD
	case errors.Is(err, subnet.ErrInvalidIP):
		reason = antibfv1.CheckV1Response_INVALID_IP
	}

	return nil, h.fail(ctx, err, reason.String())
}

func (h *handler) ResetV1(ctx context.Context, req *antibfv1.ResetV1Request) (*antibfv1.ResetV1Response, error) {
	err := h.service.Reset(ctx, req.GetLogin(), req.GetIp())
	if err == nil {
		return &antibfv1.ResetV1Response{}, nil
	}

	reason := antibfv1.ResetV1Response_UNKNOWN

	switch {
	case errors.Is(err, reset.ErrEmpty):
		reason = antibfv1.ResetV1Response_EMPTY_TARGET
	case errors.Is(err, subnet.ErrInvalidIP):
		reason = antibfv1.ResetV1Response_INVALID_IP
	}

	return nil, h.fail(ctx, err, reason.String())
}

func (h *handler) AddToBlacklistV1(
	ctx context.Context,
	req *antibfv1.AddToBlacklistV1Request,
) (*antibfv1.AddToBlacklistV1Response, error) {
	err := h.service.Add(ctx, netlist.Blacklist, req.GetSubnet())
	if err == nil {
		return &antibfv1.AddToBlacklistV1Response{}, nil
	}

	reason := antibfv1.AddToBlacklistV1Response_UNKNOWN
	if errors.Is(err, subnet.ErrInvalidSubnet) {
		reason = antibfv1.AddToBlacklistV1Response_INVALID_SUBNET
	}

	return nil, h.fail(ctx, err, reason.String())
}

func (h *handler) RemoveFromBlacklistV1(
	ctx context.Context,
	req *antibfv1.RemoveFromBlacklistV1Request,
) (*antibfv1.RemoveFromBlacklistV1Response, error) {
	err := h.service.Remove(ctx, netlist.Blacklist, req.GetSubnet())
	if err == nil {
		return &antibfv1.RemoveFromBlacklistV1Response{}, nil
	}

	reason := antibfv1.RemoveFromBlacklistV1Response_UNKNOWN
	if errors.Is(err, subnet.ErrInvalidSubnet) {
		reason = antibfv1.RemoveFromBlacklistV1Response_INVALID_SUBNET
	}

	return nil, h.fail(ctx, err, reason.String())
}

func (h *handler) AddToWhitelistV1(
	ctx context.Context,
	req *antibfv1.AddToWhitelistV1Request,
) (*antibfv1.AddToWhitelistV1Response, error) {
	err := h.service.Add(ctx, netlist.Whitelist, req.GetSubnet())
	if err == nil {
		return &antibfv1.AddToWhitelistV1Response{}, nil
	}

	reason := antibfv1.AddToWhitelistV1Response_UNKNOWN
	if errors.Is(err, subnet.ErrInvalidSubnet) {
		reason = antibfv1.AddToWhitelistV1Response_INVALID_SUBNET
	}

	return nil, h.fail(ctx, err, reason.String())
}

func (h *handler) RemoveFromWhitelistV1(
	ctx context.Context,
	req *antibfv1.RemoveFromWhitelistV1Request,
) (*antibfv1.RemoveFromWhitelistV1Response, error) {
	err := h.service.Remove(ctx, netlist.Whitelist, req.GetSubnet())
	if err == nil {
		return &antibfv1.RemoveFromWhitelistV1Response{}, nil
	}

	reason := antibfv1.RemoveFromWhitelistV1Response_UNKNOWN
	if errors.Is(err, subnet.ErrInvalidSubnet) {
		reason = antibfv1.RemoveFromWhitelistV1Response_INVALID_SUBNET
	}

	return nil, h.fail(ctx, err, reason.String())
}

// fail переводит ошибку сервиса в статус gRPC и кладёт в его детали google.rpc.ErrorInfo
// с причиной reason - именем значения ErrReason из ответа метода.
// Ошибка проверки запроса даёт InvalidArgument с её текстом. Остальные ошибки -
// сбои хранилищ: они пишутся в лог, а клиент получает Internal без подробностей.
func (h *handler) fail(ctx context.Context, err error, reason string) error {
	st := status.New(codes.InvalidArgument, err.Error())

	if !errors.Is(err, guard.ErrInvalidArgument) {
		method, _ := grpc.Method(ctx)
		h.logger.ErrorContext(ctx, "request failed", "method", method, "error", err)

		st = status.New(codes.Internal, "internal error")
	}

	withDetails, detailsErr := st.WithDetails(&errdetails.ErrorInfo{Reason: reason})
	if detailsErr != nil {
		return st.Err()
	}

	return withDetails.Err()
}

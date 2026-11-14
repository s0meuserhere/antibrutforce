//go:build integration

package integration

import (
	"testing"

	antibfv1 "github.com/s0meuserhere/antibrutforce/gen/antibf/v1"
)

func TestCheckErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *antibfv1.CheckV1Request
		want antibfv1.CheckV1Response_ErrReason
	}{
		{
			name: "empty login",
			req:  &antibfv1.CheckV1Request{Password: "password", Ip: "198.18.0.1"},
			want: antibfv1.CheckV1Response_EMPTY_LOGIN,
		},
		{
			name: "empty password",
			req:  &antibfv1.CheckV1Request{Login: "login", Ip: "198.18.0.1"},
			want: antibfv1.CheckV1Response_EMPTY_PASSWORD,
		},
		{
			name: "invalid ip",
			req:  &antibfv1.CheckV1Request{Login: "login", Password: "password", Ip: "198.18.0"},
			want: antibfv1.CheckV1Response_INVALID_IP,
		},
		{
			name: "ipv6",
			req:  &antibfv1.CheckV1Request{Login: "login", Password: "password", Ip: "::1"},
			want: antibfv1.CheckV1Response_INVALID_IP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CheckV1(t.Context(), tt.req)

			if got := rejected(t, err); got != tt.want.String() {
				t.Fatalf("reason = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResetErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *antibfv1.ResetV1Request
		want antibfv1.ResetV1Response_ErrReason
	}{
		{
			name: "empty target",
			req:  &antibfv1.ResetV1Request{},
			want: antibfv1.ResetV1Response_EMPTY_TARGET,
		},
		{
			name: "invalid ip",
			req:  &antibfv1.ResetV1Request{Login: "login", Ip: "bad"},
			want: antibfv1.ResetV1Response_INVALID_IP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.ResetV1(t.Context(), tt.req)

			if got := rejected(t, err); got != tt.want.String() {
				t.Fatalf("reason = %v, want %v", got, tt.want)
			}
		})
	}
}

// Все ручки списков одинаково отвечают INVALID_SUBNET на строку, которая не является IPv4-подсетью.
func TestListErrors(t *testing.T) {
	for _, subnet := range []string{"", "bad", "192.0.2.1", "2001:db8::/32"} {
		t.Run(subnet, func(t *testing.T) {
			_, err := client.AddToBlacklistV1(t.Context(), &antibfv1.AddToBlacklistV1Request{Subnet: subnet})
			if got := rejected(t, err); got != antibfv1.AddToBlacklistV1Response_INVALID_SUBNET.String() {
				t.Fatalf("add to blacklist: reason = %v", got)
			}

			_, err = client.RemoveFromBlacklistV1(t.Context(), &antibfv1.RemoveFromBlacklistV1Request{Subnet: subnet})
			if got := rejected(t, err); got != antibfv1.RemoveFromBlacklistV1Response_INVALID_SUBNET.String() {
				t.Fatalf("remove from blacklist: reason = %v", got)
			}

			_, err = client.AddToWhitelistV1(t.Context(), &antibfv1.AddToWhitelistV1Request{Subnet: subnet})
			if got := rejected(t, err); got != antibfv1.AddToWhitelistV1Response_INVALID_SUBNET.String() {
				t.Fatalf("add to whitelist: reason = %v", got)
			}

			_, err = client.RemoveFromWhitelistV1(t.Context(), &antibfv1.RemoveFromWhitelistV1Request{Subnet: subnet})
			if got := rejected(t, err); got != antibfv1.RemoveFromWhitelistV1Response_INVALID_SUBNET.String() {
				t.Fatalf("remove from whitelist: reason = %v", got)
			}
		})
	}
}

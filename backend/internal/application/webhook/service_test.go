package webhook

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAllowedWebhookDialTargets_RejectsPrivateIPAtDialTime(t *testing.T) {
	targets, err := resolveAllowedWebhookDialTargets(context.Background(), "example.com", "443", func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, nil
	})

	require.Error(t, err)
	assert.Nil(t, targets)
	assert.Contains(t, err.Error(), "プライベート/ループバックIPアドレス")
}

func TestResolveAllowedWebhookDialTargets_AllowsPublicIPs(t *testing.T) {
	targets, err := resolveAllowedWebhookDialTargets(context.Background(), "example.com", "443", func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}, nil
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"93.184.216.34:443"}, targets)
}

func TestResolveAllowedWebhookDialTargets_RejectsCarrierGradeNAT(t *testing.T) {
	targets, err := resolveAllowedWebhookDialTargets(context.Background(), "example.com", "443", func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("100.64.0.1")}}, nil
	})

	require.Error(t, err)
	assert.Nil(t, targets)
	assert.Contains(t, err.Error(), "プライベート/ループバックIPアドレス")
}

func TestResolveAllowedWebhookDialTargets_RejectsBenchmarkNetwork(t *testing.T) {
	targets, err := resolveAllowedWebhookDialTargets(context.Background(), "example.com", "443", func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("198.18.0.1")}}, nil
	})

	require.Error(t, err)
	assert.Nil(t, targets)
	assert.Contains(t, err.Error(), "プライベート/ループバックIPアドレス")
}

package agent

import (
	"testing"
	"context"

	"github.com/stretchr/testify/require"

	"isc.org/stork/api"
	"isc.org/stork"
)

type FakeServiceMonitor struct {
}

func (fsm *FakeServiceMonitor) GetServices() []interface{} {
	return nil
}

func (fsm *FakeServiceMonitor) Shutdown() {
}


func TestGetState(t *testing.T) {
	sa := StorkAgent{
		ServiceMonitor: &FakeServiceMonitor{},
	}

	ctx := context.Background()
	rsp, err := sa.GetState(ctx, &agentapi.GetStateReq{})
	require.NoError(t, err)
	require.Equal(t, rsp.AgentVersion, stork.Version)
}

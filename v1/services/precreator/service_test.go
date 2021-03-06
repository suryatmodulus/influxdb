package precreator_test

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/influxdb/v2/internal"
	"github.com/influxdata/influxdb/v2/toml"
	"github.com/influxdata/influxdb/v2/v1/services/precreator"
	"go.uber.org/zap/zaptest"
)

func TestShardPrecreation(t *testing.T) {
	done := make(chan struct{})
	precreate := false

	var mc internal.MetaClientMock
	mc.PrecreateShardGroupsFn = func(now, cutoff time.Time) error {
		if !precreate {
			close(done)
			precreate = true
		}
		return nil
	}

	s := NewTestService(t)
	s.MetaClient = &mc

	if err := s.Open(context.Background()); err != nil {
		t.Fatalf("unexpected open error: %s", err)
	}
	defer s.Close() // double close should not cause a panic

	timer := time.NewTimer(100 * time.Millisecond)
	select {
	case <-done:
		timer.Stop()
	case <-timer.C:
		t.Errorf("timeout exceeded while waiting for precreate")
	}

	if err := s.Close(); err != nil {
		t.Fatalf("unexpected close error: %s", err)
	}
}

func NewTestService(tb testing.TB) *precreator.Service {
	tb.Helper()

	config := precreator.NewConfig()
	config.CheckInterval = toml.Duration(10 * time.Millisecond)

	s := precreator.NewService(config)
	s.WithLogger(zaptest.NewLogger(tb))
	return s
}

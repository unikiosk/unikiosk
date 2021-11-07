package eventer

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/models"
	"github.com/unikiosk/unikiosk/pkg/util/logger"
)

func TestEventer(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.GetLoggerInstance("", zap.DebugLevel)

	e := New(ctx, log)

	events := []*models.Event{
		{
			Type: models.EventTypeProxyUpdate,
			Payload: models.KioskState{
				Content: "foo",
			},
		},
		{
			Type: models.EventTypeProxyUpdate,
			Payload: models.KioskState{
				Content: "foo1",
			},
		},
		{
			Type: models.EventTypeProxyUpdate,
			Payload: models.KioskState{
				Content: "foo2",
			},
		},
	}

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	consumer1, err := e.Subscribe(ctx1)
	require.NoError(err)
	ctx2 := (context.Background())
	consumer2, err := e.Subscribe(ctx2)
	require.NoError(err)

	var buffer1, buffer2 []*models.Event
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		var i int
		for event := range consumer1 {
			i++
			buffer1 = append(buffer1, event)
			if i == len(events) {
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		var i int
		for event := range consumer2 {
			i++
			buffer2 = append(buffer2, event)
			if i == len(events) {
				return
			}
		}
	}()

	for _, event := range events {
		require.NoError(e.Emit(event))
	}

	wg.Wait()
	require.Exactly(events, buffer1)
	require.Exactly(events, buffer2)
}

func TestEventer_iterateConsumers(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	e := &ChannelEventer{
		events: make(chan *models.Event),
		ctx:    context.Background(),
		log:    logger.GetLoggerInstance("", zap.DebugLevel),
	}

	ev := &models.Event{
		Type: models.EventTypeProxyUpdate,
		Payload: models.KioskState{
			Content: "foo",
		},
	}

	ctx1, cancel1 := context.WithCancel(context.Background())
	consumer, err := e.Subscribe(ctx1)
	require.NoError(err)
	require.Equal(1, len(e.consumers))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ev1, ok := <-consumer
		require.Exactly(ev, ev1)
		require.True(ok)
	}()
	e.iterateConsumers(ev)
	wg.Wait()

	go func() {
		cancel1()
		e.iterateConsumers(ev)
	}()
	ev1, ok := <-consumer
	require.False(ok)
	require.Nil(ev1)
	require.Equal(0, len(e.consumers))
}
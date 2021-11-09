package eventer

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
	"github.com/unikiosk/unikiosk/pkg/util/logger"
)

func TestEventer(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.GetLoggerInstance("", zap.DebugLevel)

	e := New(ctx, log)

	events := []EventWrapper{
		{
			Payload: api.Event{
				Type: api.EventTypeProxyUpdate,
				Request: api.KioskRequest{
					Content: "foo",
				},
			},
		},
		{
			Payload: api.Event{
				Type: api.EventTypeProxyUpdate,
				Request: api.KioskRequest{
					Content: "foo1",
				},
			},
		},
		{
			Payload: api.Event{
				Type: api.EventTypeProxyUpdate,
				Request: api.KioskRequest{
					Content: "foo2",
				},
			},
		},
	}

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	consumer1 := e.Subscribe(ctx1)
	ctx2 := (context.Background())
	consumer2 := e.Subscribe(ctx2)

	var buffer1, buffer2 []EventWrapper
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		var i int
		for event := range consumer1 {
			i++
			buffer1 = append(buffer1, *event)
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
			buffer2 = append(buffer2, *event)
			if i == len(events) {
				return
			}
		}
	}()

	for _, event := range events {
		go e.Emit(&event)
	}

	wg.Wait()
	require.Exactly(len(events), len(buffer1))
	require.Exactly(len(events), len(buffer2))
}

func TestEventer_iterateConsumers(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	e := &ChannelEventer{
		events: make(chan *EventWrapper),
		ctx:    context.Background(),
		log:    logger.GetLoggerInstance("", zap.DebugLevel),
	}

	ev := &EventWrapper{
		Payload: api.Event{
			Type: api.EventTypeProxyUpdate,
			Request: api.KioskRequest{
				Content: "foo",
			},
		},
	}

	ctx1, cancel1 := context.WithCancel(context.Background())
	consumer := e.Subscribe(ctx1)
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

func TestEventer_callaback(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log := logger.GetLoggerInstance("", zap.DebugLevel)

	e := New(ctx, log)

	events := []EventWrapper{
		{
			Payload: api.Event{
				Type: api.EventTypeProxyUpdate,
				Request: api.KioskRequest{
					Content: "foo",
				},
			},
		},
	}

	response := &EventWrapper{
		Payload: api.Event{
			Response: api.KioskResponse{
				Content: "bar",
			},
		},
	}

	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()
	consumer := e.Subscribe(ctx1)

	var buffer []EventWrapper
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var i int
		for event := range consumer {
			i++
			event.Callback <- response
			buffer = append(buffer, *event)
			if i == len(events) {
				return
			}
		}
	}()

	for _, event := range events {
		callback, err := e.Emit(&event)
		require.NoError(err)
		require.Exactly(callback, response)
	}

	wg.Wait()
	require.Exactly(len(events), len(buffer))
}

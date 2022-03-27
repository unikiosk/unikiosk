package eventer

// eventer is extended version of go routines, enabling multiple subscribers to same producer.

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/unikiosk/unikiosk/pkg/api"
)

var (
	// DefaultSendEventTimeout is the timeout used when publishing events to consumers
	DefaultSendEventTimeout = 2 * time.Second

	// ConsumerGCInterval is the interval at which garbage collection of consumers
	// occurs
	ConsumerGCInterval = time.Minute

	// DefaultCallabackTimeout is timeout how long emitter will wait for callaback to produce result
	DefaultCallabackTimeout = 5 * time.Second
)

var _ Eventer = &ChannelEventer{}

type EventWrapper struct {
	Payload  api.Event
	Callback chan *EventWrapper
}

type Eventer interface {
	Subscribe(ctx context.Context) <-chan *EventWrapper
	Emit(event *EventWrapper) (*EventWrapper, error)
}

// ChannelEventer is a utility to control broadcast of Events to multiple consumers.
type ChannelEventer struct {
	// events is a channel were events to be broadcasted are sent
	// This channel is never closed, because it's lifetime is tied to the
	// life of the driver and closing creates some subtile race conditions
	// between closing it and emitting events.
	events chan *EventWrapper

	// consumers is a slice of eventConsumers to broadcast events to.
	// access is guarded by consumersLock RWMutex
	consumers     []*eventConsumer
	consumersLock sync.RWMutex

	// ctx to allow control of event loop shutdown
	ctx context.Context
	log *zap.Logger
}

type eventConsumer struct {
	timeout time.Duration
	ctx     context.Context
	ch      chan *EventWrapper
	log     *zap.Logger
}

// New returns an Eventer with a running event loop that can be stopped
// by closing the given stop channel
func New(ctx context.Context, log *zap.Logger) *ChannelEventer {
	e := &ChannelEventer{
		events: make(chan *EventWrapper),
		ctx:    ctx,
		log:    log,
	}
	go e.eventLoop()
	return e
}

// eventLoop is the main logic which pulls events from the channel and broadcasts
// them to all consumers
func (e *ChannelEventer) eventLoop() {
	for {
		select {
		case <-e.ctx.Done():
			e.log.Debug("task event loop shutdown")
			return
		case event := <-e.events:
			e.iterateConsumers(event)
		case <-time.After(ConsumerGCInterval):
			e.gcConsumers()
		}
	}
}

// iterateConsumers will iterate through all consumers and broadcast the event,
// cleaning up any consumers that have closed their context
func (e *ChannelEventer) iterateConsumers(event *EventWrapper) {
	e.consumersLock.Lock()
	filtered := e.consumers[:0]
	for _, consumer := range e.consumers {

		// prioritize checking if context is cancelled prior
		// to attempting to forwarding events
		// golang select evaluations aren't predictable
		if consumer.ctx.Err() != nil {
			close(consumer.ch)
			continue
		}

		select {
		case <-time.After(consumer.timeout):
			filtered = append(filtered, consumer)
			e.log.Warn("timeout sending event", zap.Any("event", event.Payload))
		case <-consumer.ctx.Done():
			// consumer context finished, filtering it out of loop
			close(consumer.ch)
		case consumer.ch <- event:
			filtered = append(filtered, consumer)
		}
	}
	e.consumers = filtered
	e.consumersLock.Unlock()
}

func (e *ChannelEventer) gcConsumers() {
	e.consumersLock.Lock()
	filtered := e.consumers[:0]
	for _, consumer := range e.consumers {
		select {
		case <-consumer.ctx.Done():
			// consumer context finished, filtering it out of loop
		default:
			filtered = append(filtered, consumer)
		}
	}
	e.consumers = filtered
	e.consumersLock.Unlock()
}

func (e *ChannelEventer) newConsumer(ctx context.Context) *eventConsumer {
	e.consumersLock.Lock()
	defer e.consumersLock.Unlock()

	consumer := &eventConsumer{
		ch:      make(chan *EventWrapper),
		ctx:     ctx,
		timeout: DefaultSendEventTimeout,
		log:     e.log,
	}
	e.consumers = append(e.consumers, consumer)

	return consumer
}

// Subscribe subscribes to events
func (e *ChannelEventer) Subscribe(ctx context.Context) <-chan *EventWrapper {
	consumer := e.newConsumer(ctx)
	return consumer.ch
}

// Emit emits event to all subscribers
func (e *ChannelEventer) Emit(event *EventWrapper) (*EventWrapper, error) {
	if event.Callback == nil {
		event.Callback = make(chan *EventWrapper)
	}
	callbackTimout := time.NewTimer(DefaultCallabackTimeout)

	select {
	case <-e.ctx.Done():
		return nil, e.ctx.Err()
	case <-callbackTimout.C:
		return nil, fmt.Errorf("callback timeout")
	case e.events <- event:
		e.log.Warn("emitting event", zap.Any("event", event.Payload))
	}

	select {
	case <-e.ctx.Done():
		return nil, e.ctx.Err()
	case <-callbackTimout.C:
		return nil, fmt.Errorf("callback timeout")
	case v := <-event.Callback:
		return v, nil
	}

}

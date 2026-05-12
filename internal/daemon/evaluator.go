package daemon

import (
	"context"
	"time"

	"github.com/beaallombert/gotask/internal/domain"
	"github.com/beaallombert/gotask/internal/notifications"
	"github.com/beaallombert/gotask/internal/rules"
)

type StateProvider func(context.Context) (*domain.SystemState, error)
type InterventionSink func(context.Context, *domain.Intervention)

// Evaluator periodically evaluates rules and emits at most one intervention per cycle.
type Evaluator struct {
	engine      *rules.Engine
	notifier    notifications.Notifier
	interval    time.Duration
	onIntervene InterventionSink
	onError     func(error)
}

func NewEvaluator(engine *rules.Engine, notifier notifications.Notifier, interval time.Duration) *Evaluator {
	if engine == nil {
		engine = rules.NewEngine()
	}
	if notifier == nil {
		notifier = notifications.NewNotifier()
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}

	return &Evaluator{
		engine:   engine,
		notifier: notifier,
		interval: interval,
	}
}

func (e *Evaluator) SetInterventionSink(sink InterventionSink) {
	e.onIntervene = sink
}

func (e *Evaluator) SetErrorHandler(handler func(error)) {
	e.onError = handler
}

func (e *Evaluator) Run(ctx context.Context, provider StateProvider) {
	if provider == nil {
		return
	}

	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		if !e.evaluateOnce(ctx, provider) {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (e *Evaluator) evaluateOnce(ctx context.Context, provider StateProvider) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	state, err := provider(ctx)
	if err != nil {
		if e.onError != nil {
			e.onError(err)
		}
		return true
	}

	interventions := e.engine.Evaluate(state)
	if len(interventions) == 0 {
		return true
	}

	chosen := interventions[0]
	_ = e.notifier.Send("gotask", chosen.Message)
	if e.onIntervene != nil {
		e.onIntervene(ctx, chosen)
	}

	return true
}

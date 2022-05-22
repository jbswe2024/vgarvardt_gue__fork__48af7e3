package gue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/nonrecording"
	"go.opentelemetry.io/otel/trace"

	"github.com/vgarvardt/gue/v4/adapter"
)

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string, fields ...adapter.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Info(msg string, fields ...adapter.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Error(msg string, fields ...adapter.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) With(fields ...adapter.Field) adapter.Logger {
	args := m.Called(fields)
	return args.Get(0).(adapter.Logger)
}

var dummyWM = WorkMap{
	"MyJob": func(ctx context.Context, j *Job) error {
		return nil
	},
}

func TestWithWorkerPollInterval(t *testing.T) {
	workerWithDefaultInterval, err := NewWorker(nil, dummyWM)
	require.NoError(t, err)
	assert.Equal(t, defaultPollInterval, workerWithDefaultInterval.interval)

	customInterval := 12345 * time.Millisecond
	workerWithCustomInterval, err := NewWorker(nil, dummyWM, WithWorkerPollInterval(customInterval))
	require.NoError(t, err)
	assert.Equal(t, customInterval, workerWithCustomInterval.interval)
}

func TestWithWorkerQueue(t *testing.T) {
	workerWithDefaultQueue, err := NewWorker(nil, dummyWM)
	require.NoError(t, err)
	assert.Equal(t, defaultQueueName, workerWithDefaultQueue.queue)

	customQueue := "fooBarBaz"
	workerWithCustomQueue, err := NewWorker(nil, dummyWM, WithWorkerQueue(customQueue))
	require.NoError(t, err)
	assert.Equal(t, customQueue, workerWithCustomQueue.queue)
}

func TestWithWorkerID(t *testing.T) {
	workerWithDefaultID, err := NewWorker(nil, dummyWM)
	require.NoError(t, err)
	assert.NotEmpty(t, workerWithDefaultID.id)

	customID := "some-meaningful-id"
	workerWithCustomID, err := NewWorker(nil, dummyWM, WithWorkerID(customID))
	require.NoError(t, err)
	assert.Equal(t, customID, workerWithCustomID.id)
}

func TestWithWorkerLogger(t *testing.T) {
	workerWithDefaultLogger, err := NewWorker(nil, dummyWM)
	require.NoError(t, err)
	assert.IsType(t, adapter.NoOpLogger{}, workerWithDefaultLogger.logger)

	logMessage := "hello"

	l := new(mockLogger)
	l.On("Info", logMessage, mock.Anything)
	// worker sets id as default logger field
	l.On("With", mock.Anything).Return(l)

	workerWithCustomLogger, err := NewWorker(nil, dummyWM, WithWorkerLogger(l))
	require.NoError(t, err)
	workerWithCustomLogger.logger.Info(logMessage)

	l.AssertExpectations(t)
}

func TestWithWorkerPollStrategy(t *testing.T) {
	workerWithWorkerPollStrategy, err := NewWorker(nil, dummyWM, WithWorkerPollStrategy(RunAtPollStrategy))
	require.NoError(t, err)
	assert.Equal(t, RunAtPollStrategy, workerWithWorkerPollStrategy.pollStrategy)
}

func TestWithPoolPollInterval(t *testing.T) {
	workerPoolWithDefaultInterval, err := NewWorkerPool(nil, dummyWM, 2)
	require.NoError(t, err)
	assert.Equal(t, defaultPollInterval, workerPoolWithDefaultInterval.interval)

	customInterval := 12345 * time.Millisecond
	workerPoolWithCustomInterval, err := NewWorkerPool(nil, dummyWM, 2, WithPoolPollInterval(customInterval))
	require.NoError(t, err)
	assert.Equal(t, customInterval, workerPoolWithCustomInterval.interval)
}

func TestWithPoolQueue(t *testing.T) {
	workerPoolWithDefaultQueue, err := NewWorkerPool(nil, dummyWM, 2)
	require.NoError(t, err)
	assert.Equal(t, defaultQueueName, workerPoolWithDefaultQueue.queue)

	customQueue := "fooBarBaz"
	workerPoolWithCustomQueue, err := NewWorkerPool(nil, dummyWM, 2, WithPoolQueue(customQueue))
	require.NoError(t, err)
	assert.Equal(t, customQueue, workerPoolWithCustomQueue.queue)
}

func TestWithPoolID(t *testing.T) {
	workerPoolWithDefaultID, err := NewWorkerPool(nil, dummyWM, 2)
	require.NoError(t, err)
	assert.NotEmpty(t, workerPoolWithDefaultID.id)

	customID := "some-meaningful-id"
	workerPoolWithCustomID, err := NewWorkerPool(nil, dummyWM, 2, WithPoolID(customID))
	require.NoError(t, err)
	assert.Equal(t, customID, workerPoolWithCustomID.id)
}

func TestWithPoolLogger(t *testing.T) {
	workerPoolWithDefaultLogger, err := NewWorkerPool(nil, dummyWM, 2)
	require.NoError(t, err)
	assert.IsType(t, adapter.NoOpLogger{}, workerPoolWithDefaultLogger.logger)

	logMessage := "hello"

	l := new(mockLogger)
	l.On("Info", logMessage, mock.Anything)
	// worker pool sets id as default logger field
	l.On("With", mock.Anything).Return(l)

	workerPoolWithCustomLogger, err := NewWorkerPool(nil, dummyWM, 2, WithPoolLogger(l))
	require.NoError(t, err)
	workerPoolWithCustomLogger.logger.Info(logMessage)

	l.AssertExpectations(t)
}

func TestWithPoolPollStrategy(t *testing.T) {
	workerPoolWithPoolPollStrategy, err := NewWorkerPool(nil, dummyWM, 2, WithPoolPollStrategy(RunAtPollStrategy))
	require.NoError(t, err)
	assert.Equal(t, RunAtPollStrategy, workerPoolWithPoolPollStrategy.pollStrategy)
}

func TestWithPoolTracer(t *testing.T) {
	customTracer := trace.NewNoopTracerProvider().Tracer("custom")

	workerPoolWithTracer, err := NewWorkerPool(nil, dummyWM, 2, WithPoolTracer(customTracer))
	require.NoError(t, err)
	assert.Equal(t, customTracer, workerPoolWithTracer.tracer)

	for i := range workerPoolWithTracer.workers {
		assert.Equal(t, customTracer, workerPoolWithTracer.workers[i].tracer)
	}
}

func TestWithPoolMeter(t *testing.T) {
	customMeter := nonrecording.NewNoopMeterProvider().Meter("custom")

	workerPoolWithMeter, err := NewWorkerPool(nil, dummyWM, 2, WithPoolMeter(customMeter))
	require.NoError(t, err)
	assert.Equal(t, customMeter, workerPoolWithMeter.meter)

	for i := range workerPoolWithMeter.workers {
		assert.Equal(t, customMeter, workerPoolWithMeter.workers[i].meter)
	}
}

type dummyHook struct {
	counter int
}

func (h *dummyHook) handler(context.Context, *Job, error) {
	h.counter++
}

func TestWithWorkerHooksJobLocked(t *testing.T) {
	ctx := context.Background()
	hook := new(dummyHook)

	workerWOutHooks, err := NewWorker(nil, dummyWM)
	require.NoError(t, err)
	for _, h := range workerWOutHooks.hooksJobLocked {
		h(ctx, nil, nil)
	}
	require.Equal(t, 0, hook.counter)

	workerWithHooks, err := NewWorker(nil, dummyWM, WithWorkerHooksJobLocked(hook.handler, hook.handler, hook.handler))
	require.NoError(t, err)
	for _, h := range workerWithHooks.hooksJobLocked {
		h(ctx, nil, nil)
	}
	require.Equal(t, 3, hook.counter)
}

func TestWithWorkerHooksUnknownJobType(t *testing.T) {
	ctx := context.Background()
	hook := new(dummyHook)

	workerWOutHooks, err := NewWorker(nil, dummyWM)
	require.NoError(t, err)
	for _, h := range workerWOutHooks.hooksUnknownJobType {
		h(ctx, nil, nil)
	}
	require.Equal(t, 0, hook.counter)

	workerWithHooks, err := NewWorker(nil, dummyWM, WithWorkerHooksUnknownJobType(hook.handler, hook.handler, hook.handler))
	require.NoError(t, err)
	for _, h := range workerWithHooks.hooksUnknownJobType {
		h(ctx, nil, nil)
	}
	require.Equal(t, 3, hook.counter)
}

func TestWithWorkerHooksJobDone(t *testing.T) {
	ctx := context.Background()
	hook := new(dummyHook)

	workerWOutHooks, err := NewWorker(nil, dummyWM)
	require.NoError(t, err)
	for _, h := range workerWOutHooks.hooksJobDone {
		h(ctx, nil, nil)
	}
	require.Equal(t, 0, hook.counter)

	workerWithHooks, err := NewWorker(nil, dummyWM, WithWorkerHooksJobDone(hook.handler, hook.handler, hook.handler))
	require.NoError(t, err)
	for _, h := range workerWithHooks.hooksJobDone {
		h(ctx, nil, nil)
	}
	require.Equal(t, 3, hook.counter)
}

func TestWithPoolHooksJobLocked(t *testing.T) {
	ctx := context.Background()
	hook := new(dummyHook)

	poolWOutHooks, err := NewWorkerPool(nil, dummyWM, 3)
	require.NoError(t, err)
	for _, w := range poolWOutHooks.workers {
		for _, h := range w.hooksJobLocked {
			h(ctx, nil, nil)
		}
	}
	require.Equal(t, 0, hook.counter)

	poolWithHooks, err := NewWorkerPool(nil, dummyWM, 3, WithPoolHooksJobLocked(hook.handler, hook.handler, hook.handler))
	require.NoError(t, err)
	for _, w := range poolWithHooks.workers {
		for _, h := range w.hooksJobLocked {
			h(ctx, nil, nil)
		}
	}
	require.Equal(t, 9, hook.counter)
}

func TestWithPoolHooksUnknownJobType(t *testing.T) {
	ctx := context.Background()
	hook := new(dummyHook)

	poolWOutHooks, err := NewWorkerPool(nil, dummyWM, 3)
	require.NoError(t, err)
	for _, w := range poolWOutHooks.workers {
		for _, h := range w.hooksUnknownJobType {
			h(ctx, nil, nil)
		}
	}
	require.Equal(t, 0, hook.counter)

	poolWithHooks, err := NewWorkerPool(nil, dummyWM, 3, WithPoolHooksUnknownJobType(hook.handler, hook.handler, hook.handler))
	require.NoError(t, err)
	for _, w := range poolWithHooks.workers {
		for _, h := range w.hooksUnknownJobType {
			h(ctx, nil, nil)
		}
	}
	require.Equal(t, 9, hook.counter)
}

func TestWithPoolHooksJobDone(t *testing.T) {
	ctx := context.Background()
	hook := new(dummyHook)

	poolWOutHooks, err := NewWorkerPool(nil, dummyWM, 3)
	require.NoError(t, err)
	for _, w := range poolWOutHooks.workers {
		for _, h := range w.hooksJobDone {
			h(ctx, nil, nil)
		}
	}
	require.Equal(t, 0, hook.counter)

	poolWithHooks, err := NewWorkerPool(nil, dummyWM, 3, WithPoolHooksJobDone(hook.handler, hook.handler, hook.handler))
	require.NoError(t, err)
	for _, w := range poolWithHooks.workers {
		for _, h := range w.hooksJobDone {
			h(ctx, nil, nil)
		}
	}
	require.Equal(t, 9, hook.counter)
}

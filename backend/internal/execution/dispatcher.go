package execution

import "context"

type Dispatcher interface {
	Dispatch(ctx context.Context, task *Task) error
}

type SyncDispatcher struct {
	engine Engine
}

func NewSyncDispatcher(engine Engine) *SyncDispatcher {
	return &SyncDispatcher{engine: engine}
}

func (d *SyncDispatcher) Dispatch(ctx context.Context, task *Task) error {
	result := d.engine.Execute(ctx, task)
	return result.Error
}

type AsyncDispatcher struct {
	engine Engine
}

func NewAsyncDispatcher(engine Engine) *AsyncDispatcher {
	return &AsyncDispatcher{engine: engine}
}

func (d *AsyncDispatcher) Dispatch(ctx context.Context, task *Task) error {
	go func() {
		d.engine.Execute(ctx, task)
	}()
	return nil
}

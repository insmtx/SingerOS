package worker

import (
	"github.com/insmtx/SingerOS/backend/internal/worker/client"
)

type Worker = client.Worker
type WorkerConfig = client.WorkerConfig

var NewWorker = client.NewWorker

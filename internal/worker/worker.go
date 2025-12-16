package worker

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Worker struct {
	rdb *redis.Client
}

func NewWorker(rdb *redis.Client) *Worker {
	return &Worker{rdb: rdb}
}

// StartCron initiates the background cron job with distributed locking
func (w *Worker) StartCron(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	lockKey := "cron:leader"

	// Simple unique ID for this instance
	// In production, use pod ID or hostname
	instanceID := "instance-" + time.Now().String()

	go func() {
		for {
			select {
			case <-ticker.C:
				w.tryLockAndRun(ctx, lockKey, instanceID)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (w *Worker) tryLockAndRun(ctx context.Context, lockKey, instanceID string) {
	// 1. Try to acquire lock (10s TTL)
	acquired, err := w.rdb.SetNX(ctx, lockKey, instanceID, 10*time.Second).Result()
	if err != nil {
		log.Printf("worker lock error: %v", err)
		return
	}

	// 2. If acquired (or we already hold it), renew lease and work
	if acquired || w.rdb.Get(ctx, lockKey).Val() == instanceID {
		w.rdb.Expire(ctx, lockKey, 10*time.Second) // Renew lease

		// Do the work
		w.runCleanupJob()
	}
}

func (w *Worker) runCleanupJob() {
	log.Println("Cron job running: cleanup...")
	// Add actual cleanup logic here
}

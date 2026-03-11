package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ListenForNotifications listens for PG notifications on a channel and calls the handler
func ListenForNotifications(ctx context.Context, pool *pgxpool.Pool, channel string, handler func(payload []byte)) {
	for {
		err := listen(ctx, pool, channel, handler)
		if err != nil {
			log.Printf("DB Listener error: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func listen(ctx context.Context, pool *pgxpool.Pool, channel string, handler func(payload []byte)) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN "+channel)
	if err != nil {
		return err
	}

	log.Printf("Listening for notifications on channel: %s", channel)

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}

		handler([]byte(notification.Payload))
	}
}

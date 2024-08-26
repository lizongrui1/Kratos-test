package server

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"student/internal/data"
	"time"
)

func StartStudentEventConsumer(rdb data.RedisClientRepo) {
	go func() {
		for {
			result, err := rdb.PopMsg(context.Background(), 2*time.Second, "student:add", "student:delete", "student:update").Result()
			if err != nil {
				log.Errorf("Error fetching event from Redis queue: %v", err)
				continue
			}

			eventType := result[0]
			studentID := result[1]

			switch eventType {
			case "student:add":
				handleStudentAdd(studentID)
			case "student:delete":
				handleStudentDelete(studentID)
			case "student:update":
				handleStudentUpdate(studentID)
			}
		}
	}()
	log.Infof("Redis consumer started")
}

// 示例处理函数
func handleStudentAdd(studentID string) {
	log.Infof("Handling student add event for ID: %s", studentID)
	// 处理逻辑
}

func handleStudentDelete(studentID string) {
	log.Infof("Handling student delete event for ID: %s", studentID)
	// 处理逻辑
}

func handleStudentUpdate(studentID string) {
	log.Infof("Handling student update event for ID: %s", studentID)
	// 处理逻辑
}

package tasks

import (
	"ASO/main/database"
	"context"
	"fmt"
	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var Scheduler *gocron.Scheduler

func StartRepeatingTasks() {
	Scheduler = gocron.NewScheduler(time.UTC)

	//startCheckerTask()

	Scheduler.StartAsync()

	fmt.Println("Started scheduler")
}

func StopRepeatingTasks() {
	Scheduler.Clear()
}

func startCheckerTask() {
	_, err := Scheduler.Every(1).Hour().Do(func() {
		fmt.Println("Checking Groups Users and Tokens")

		database.MongoDB.Collection("userGroup").FindOne(context.Background(), bson.M{
			"expires":  true,
			"notify":   true,
			"notified": false,
			"dateExpires": bson.M{
				"$lte": time.Now(),
			},
		})

	})
	if err != nil {
		fmt.Println("Failed to start checker task")
		fmt.Println(err)
	}
}

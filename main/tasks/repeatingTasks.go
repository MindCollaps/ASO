package tasks

import (
	"github.com/go-co-op/gocron"
	"log"
	"time"
)

var Scheduler *gocron.Scheduler

func StartRepeatingTasks() {
	Scheduler = gocron.NewScheduler(time.UTC)

	initCheckerTasks()

	Scheduler.StartAsync()

	log.Println("Initialized scheduler")
}

func StopRepeatingTasks() {
	Scheduler.Clear()
}

func updateGitUserState() {

}

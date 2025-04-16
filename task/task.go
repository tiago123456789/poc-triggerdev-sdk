package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	httpclient "github.com/tiago123456789/poc-triggerdev-sdk/http-client"
	"github.com/tiago123456789/poc-triggerdev-sdk/logger"

	"github.com/adhocore/gronx"
)

type ITask interface {
	Add(TaskScheduled) error
	Execute(id string, data map[string]interface{})
	Start() error
}

type ActionFunction func(message map[string]interface{}, logger *slog.Logger) error

type TaskScheduled struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Cron   string `json:"cron"`
	Action ActionFunction
}

type TaskScheduledToRegister struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Cron         string `json:"cron"`
	UrlToTrigger string `json:"url_to_trigger"`
}

type Task struct {
	tasksScheduled map[string]TaskScheduled
}

func (t *Task) notifyFinished(id string) {
	jsonData, err := json.Marshal(map[string]interface{}{})

	url := os.Getenv("REMOTE_TRIGGER_ENDPOINT")
	url = fmt.Sprintf("%s-finished-execution/%s", url, id)
	err = httpclient.PostRequest(url, jsonData)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (t *Task) Execute(id string, data map[string]interface{}) {
	task := t.tasksScheduled[id]
	logger := logger.Init([]slog.Attr{
		{
			Key: "id", Value: slog.StringValue(task.Id),
		},
		{
			Key: "name", Value: slog.StringValue(task.Name),
		},
	})

	logger.Info("Starting execution", "id", id)
	err := task.Action(data, logger)
	if err == nil {
		logger.Info("Finished execution with success")
	} else {
		logger.Error(
			fmt.Sprintf("Finished execution with error: %s", err.Error()),
		)
	}

	t.notifyFinished(id)
}

func (t *Task) Add(item TaskScheduled) error {
	if gronx.IsValid(item.Cron) == false {
		panic("Cron expression invalid.")
	}

	t.tasksScheduled[item.Id] = item

	return nil
}

func (t *Task) Start() error {
	items := []TaskScheduledToRegister{}

	for _, item := range t.tasksScheduled {
		items = append(items, TaskScheduledToRegister{
			Id:           item.Id,
			Name:         item.Name,
			Cron:         item.Cron,
			UrlToTrigger: os.Getenv("URL_TO_TRIGGER"),
		})
	}
	jsonData, err := json.Marshal(items)
	if err != nil {
		panic(err)
	}

	url := os.Getenv("REMOTE_TRIGGER_ENDPOINT")
	err = httpclient.PostRequest(url, jsonData)
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func Init() *Task {
	if os.Getenv("REMOTE_TRIGGER_ENDPOINT") == "" {
		panic("You forgot set the env REMOTE_TRIGGER_ENDPOINT")
	}

	if os.Getenv("URL_TO_TRIGGER") == "" {
		panic("You forgot set the env URL_TO_TRIGGER")
	}

	if os.Getenv("REMOTE_TRIGGER_LOGGERS_ENDPOINT") == "" {
		panic("You forgot to set the env REMOTE_TRIGGER_LOGGERS_ENDPOINT")
	}

	emptyList := map[string]TaskScheduled{}
	return &Task{
		tasksScheduled: emptyList,
	}
}

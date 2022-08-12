package job

import (
	"github.com/robfig/cron/v3"
	"kube-terminal/k8s-webshell/client"
)

var TaskCron *cron.Cron

func InitAndStart() {
	TaskCron = cron.New()
	defer TaskCron.Stop().Done()
	_, err := TaskCron.AddFunc("0/5 * * * * ? *", client.CheckTokenTimeout)
	if err != nil {
		return
	}
	// start
	TaskCron.Start()
}

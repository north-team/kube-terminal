package job

import (
	"kube-terminal/client"
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

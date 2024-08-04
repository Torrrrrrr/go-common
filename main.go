package main

import (
	applogger "common-service/pkg/appLogger"
	"time"
)

func main() {

	log := applogger.NewAppLogger("DEBUG").WithField(applogger.AppName, "TestApp111")
	log.Info("[INFO]", time.Now().UnixNano(), "main", "Hello world!")
	log.Error("[ERROR]", time.Now().UnixNano(), "main", 1024)

	log2 := log.Clone().WithField(applogger.AppName, "TestApp222")
	log2.Info("[INFO]", time.Now().UnixNano(), "main", log2)
	log2.Debug("[DEBUG]", time.Now().UnixNano(), "main", 555)

}

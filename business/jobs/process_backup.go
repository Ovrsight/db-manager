package jobs

import (
	"github.com/nizigama/ovrsight/foundation/backup"
	"github.com/nizigama/ovrsight/foundation/storage"
	"log"
	"sync"
)

type BackupProcessor struct {
}

func (bp *BackupProcessor) ProcessBackup(method backup.Method, engine storage.Engine) bool {

	wg := sync.WaitGroup{}
	backupSuccessful := false

	wg.Add(2)
	dataChan := make(chan []byte)

	go func() {
		// backup method cleaner
		defer method.Clean(dataChan)
		defer wg.Done()

		// generate backup bytes
		err := method.Generate(dataChan)

		if err != nil {
			log.Fatalln("Backup failure:", err)
		}
	}()

	go func(se storage.Engine) {

		defer wg.Done()

		// upload backup bytes
		err := se.Save(dataChan)

		if err != nil {
			log.Fatalln("Storage failure:", err)
		}
		backupSuccessful = true
	}(engine)

	wg.Wait()

	return backupSuccessful
}

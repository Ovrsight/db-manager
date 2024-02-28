package jobs

import (
	"errors"
	"github.com/nizigama/ovrsight/foundation/backup"
	"github.com/nizigama/ovrsight/foundation/storage"
	"log"
	"sync"
)

type BackupProcessor struct {
}

func (bp *BackupProcessor) ProcessBackup(method backup.Method, engine storage.Engine, onSuccess func(size int)) error {

	wg := sync.WaitGroup{}
	backupSuccessful := false
	generatingSuccessful := false
	uploadedSize := 0

	wg.Add(2)
	dataChan := make(chan []byte)
	failureChan := make(chan struct{})

	go func() {
		// backup method cleaner
		defer method.Clean(dataChan)
		defer wg.Done()

		// generate backup bytes
		err := method.Generate(dataChan, failureChan)
		if err != nil {
			log.Println("Backup failure:", err)
			failureChan <- struct{}{}
			return
		}
		generatingSuccessful = true
	}()

	go func(se storage.Engine) {

		defer wg.Done()

		// upload backup bytes
		uploaded, err := se.Save(dataChan, failureChan)

		if err != nil {
			log.Println("Storage failure:", err)
			failureChan <- struct{}{}
			return
		}

		backupSuccessful = true
		uploadedSize = uploaded
	}(engine)

	wg.Wait()

	if !backupSuccessful || !generatingSuccessful {
		return errors.New("failed to process the backup")
	}

	onSuccess(uploadedSize)

	return nil
}

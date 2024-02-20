package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

func (d *Database) FindOrCreate(database string) error {

	res := Db.First(&d, "name = ?", database)

	if res.Error == nil {
		return nil
	}

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {

		now := time.Now()

		// create database
		d.Name = database
		d.FirstBackupTime = now
		d.LatestBackupTime = now

		res = Db.Create(d)

		if res.Error != nil {
			return res.Error
		}

		return nil
	}

	return res.Error
}

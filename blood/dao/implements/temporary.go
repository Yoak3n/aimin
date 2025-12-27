package implements

import (
	"time"

	"blood/pkg/util"
	"blood/schema"
)

func (d *Database) CreateTemporaryTableRecord(e schema.TemporaryMemoryTable) error {
	if e.Id == "" {
		e.Id = util.RandomIdWithPrefix("temporary-")
	}
	return d.GetPostgresSQL().Create(e).Error
}

func (d *Database) QueryTemporaryTableRecord(id string) (schema.TemporaryMemoryTable, error) {
	var e schema.TemporaryMemoryTable
	return e, d.GetPostgresSQL().Where("id = ?", id).First(&e).Error
}

func (d *Database) UpdateTemporaryTableRecord(e schema.TemporaryMemoryTable) error {
	return d.GetPostgresSQL().Save(&e).Error
}

func (d *Database) QueryTemporaryTableRecords() ([]schema.TemporaryMemoryTable, error) {
	var es []schema.TemporaryMemoryTable
	now := time.Now()
	return es, d.GetPostgresSQL().Where("last_simulate_time >= ?", now.AddDate(0, 0, -7)).Find(&es).Error
}

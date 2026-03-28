package implements

import (
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
)

func (d *Database) CreateSummaryMemoryTableRecord(r *schema.SummaryMemoryTable) error {
	if r.Id == "" {
		r.Id = util.RandomIdWithPrefix("sum-")
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = r.CreatedAt
	}
	return d.GetPostgresSQL().Create(r).Error
}

func (d *Database) GetSummaryMemoryTableRecord(id string) (schema.SummaryMemoryTable, error) {
	var r schema.SummaryMemoryTable
	return r, d.GetPostgresSQL().Where("id = ?", id).First(&r).Error
}

func (d *Database) GetSummaryMemoryTableRecordByLink(link string) (schema.SummaryMemoryTable, error) {
	var r schema.SummaryMemoryTable
	return r, d.GetPostgresSQL().Where("link = ?", link).Order("updated_at desc").First(&r).Error
}

func (d *Database) ListSummaryMemoryTableRecords(limit int) ([]schema.SummaryMemoryTable, error) {
	l := 50
	if limit > 0 {
		l = limit
	}
	var rs []schema.SummaryMemoryTable
	return rs, d.GetPostgresSQL().Order("created_at desc").Limit(l).Find(&rs).Error
}

func (d *Database) UpdateSummaryMemoryTableRecord(r *schema.SummaryMemoryTable) error {
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = time.Now()
	}
	return d.GetPostgresSQL().Save(r).Error
}

func (d *Database) DeleteSummaryMemoryTableRecord(id string) error {
	return d.GetPostgresSQL().Delete(&schema.SummaryMemoryTable{}, "id = ?", id).Error
}

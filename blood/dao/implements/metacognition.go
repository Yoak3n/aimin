package implements

import "blood/schema"

func (d *Database) CreateMetacognitionRecord(r schema.MetacognitionTable) error {
	db := d.GetPostgresSQL()
	return db.Create(&r).Error
}

func (d *Database) GetMetacognitionRecord(id int) (*schema.MetacognitionTable, error) {
	db := d.GetPostgresSQL()
	var r schema.MetacognitionTable
	return &r, db.First(&r, id).Error
}

func (d *Database) GetMetacognitionRecords(limit int) ([]schema.MetacognitionTable, error) {
	l := 10
	if limit > 0 {
		l = limit
	}
	db := d.GetPostgresSQL()
	var r []schema.MetacognitionTable
	return r, db.Limit(l).Order("confidence_score desc").Find(&r).Error
}

func (d *Database) UpdateMetacognitionRecord(r schema.MetacognitionTable) error {
	db := d.GetPostgresSQL()
	return db.Save(&r).Error
}

func (d *Database) DeleteMetacognitionRecord(id int) error {
	db := d.GetPostgresSQL()
	return db.Delete(&schema.MetacognitionTable{}, id).Error
}

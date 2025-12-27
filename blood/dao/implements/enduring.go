package implements

import (
	"log"

	"blood/pkg/util"
	"blood/schema"
)

func (d *Database) CreateEnduringTableRecord(e schema.EnduringMemoryTable, embedding []float32) error {
	if e.Id == "" {
		e.Id = util.RandomIdWithPrefix("enduring-")
	}
	embeddingStr := util.Float32SliceToString(embedding)
	res := d.GetPostgresSQL().Exec(`INSERT INTO enduring_memory
		(id, content, topic, count, last_simulated_time, embedding) 
		VALUES ($1, $2, $3, $4, $5, $6);`,
		e.Id, e.Content, e.Topic, e.Count, e.LastSimulatedTime, embeddingStr)

	return res.Error
}

func (d *Database) QueryEnduringTableRecords() ([]schema.EnduringMemoryTable, error) {
	var es []schema.EnduringMemoryTable
	return es, d.GetPostgresSQL().Find(&es).Error
}

func (d *Database) AddEntityTableRecord(e []schema.EntityTable) []uint {
	d.PostgresDB.Create(e)
	ret := make([]uint, len(e))
	for index := range e {
		ret[index] = e[index].ID
	}

	go func() {
		err := d.NeuroDB.CreateNode(e)
		if err != nil {
			log.Println(err)
		}
	}()
	return ret
}

func (d *Database) QueryEnduringIDs(emID string) ([]uint, error) {
	var ret []uint
	return ret, d.GetPostgresSQL().Where("link = ?", emID).Model(&schema.EntityTable{}).Pluck("id", &ret).Error
}

func (d *Database) QueryEntityTableRecordsByLink(emID string) ([]schema.EntityTable, error) {
	var ret []schema.EntityTable
	return ret, d.GetPostgresSQL().Where("link = ?", emID).Find(&ret).Error
}

func (d *Database) QueryClosestEnduringMemoryDialogue(query []float32) []schema.DialogueRecord {
	em := d.QueryClosestEnduringMemoryTableRecords(query)
	if len(em) == 0 {
		return nil
	}
	linkedIds := make([]string, 0)
	for _, e := range em {
		linkedIds = append(linkedIds, e.Id)
	}
	return d.QueryDialogueRecordsByLinks(linkedIds)
}

func (d *Database) QueryClosestEnduringMemoryTableRecords(query []float32) []schema.EnduringMemoryTable {
	embeddingStr := util.Float32SliceToString(query)
	var ret []schema.EnduringMemoryTable
	res := d.GetPostgresSQL().Raw(`SELECT * FROM enduring_memory 
		ORDER BY embedding <-> $1::vector LIMIT 5;`, embeddingStr).Scan(&ret)
	if res.Error != nil {
		return nil
	}
	return ret
}

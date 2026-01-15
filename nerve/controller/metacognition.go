package controller

import (
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
)

func GetCurrentMetacognitionRecords() ([]schema.MetacognitionTable, error) {
	records, err := helper.UseDB().GetMetacognitionRecords(10)
	if err != nil {
		return nil, err
	}
	return records, nil
}

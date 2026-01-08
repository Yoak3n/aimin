package controller

import (
	"blood/pkg/helper"
	"blood/schema"
)

func GetCurrentMetacognitionRecords() ([]schema.MetacognitionTable, error) {
	records, err := helper.UseDB().GetMetacognitionRecords(10)
	if err != nil {
		return nil, err
	}
	return records, nil
}

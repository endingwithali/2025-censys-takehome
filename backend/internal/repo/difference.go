package repo

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Differences struct {
	gorm.Model
	Host_IP    string          `json:"host_ip"`
	Timestamp1 time.Time       `json:"timestamp1"`
	Timestamp2 time.Time       `json:"timestamp2"`
	JSON_Data  json.RawMessage `json:"json_data"`
}

/*
Explicitly telilng GORM the table name to use for snapshot differences
*/
func (Differences) TableName() string {
	return "snapshot_differences"
}

func InsertDifference(difference Differences) error {
	// ctx := context.Background()

	// err := DB.WithContext(ctx).Where(
	// 	"host_ip=? AND timestamp1=? AND timestamp2=?",
	// 	host_ip, timestamp1, timestamp2,
	// )
	// err := DB.WithContext(ctx).Create(difference).Error
	// return err
	return nil
}

func CheckForComparison(host_ip string, timestamp1 time.Time, timestamp2 time.Time) (Differences, error) {
	// ctx := context.Background()
	// var differences SnapshotDifferences
	// err := DB.WithContext(ctx).Where(
	// 	"host_ip=? AND timestamp1=? AND timestamp2=?",
	// 	host_ip, timestamp1, timestamp2,
	// ).First(&differences).Error

	// if err != nil {
	// 	return differences, err
	// }
	// return differences, nil
	return Differences{}, nil
}

package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Snapshot model used by
type Snapshot struct {
	UUID      uuid.UUID `json:"uuid" gorm:"column:uuid"`
	Host_IP   string    `json:"host_ip" gorm:"column:host_ip"`
	Timestamp time.Time `json:"timestamp" gorm:"column:timestamp"`
	File_PWD  string    `json:"file_pwd" gorm:"column:file_pwd"`
	File_Name string    `json:"file_name" gorm:"column:file_name"`
}

func (Snapshot) TableName() string {
	return "snapshot"
}

type SnapshotRepo interface {
	Insert(ctx context.Context, host_ip string, timestap time.Time, file_pwd string, file_name string) error
	GetSnapshotByTimeStamp(ctx context.Context, host_ip string, timestamp time.Time) (Snapshot, error)
	GetSnapshotByFileName(ctx context.Context, host_ip string, filename string) (Snapshot, error)
	GetAllHosts(ctx context.Context) ([]string, error)
	ListAllHostSnapshots(ctx context.Context, host_ip string) ([]string, error)
}

type snapshotRepo struct {
	db *gorm.DB
}

func NewSnapshotRepo(db *gorm.DB) SnapshotRepo {
	return &snapshotRepo{
		db: db,
	}
}

func (sr *snapshotRepo) Insert(ctx context.Context, host_ip string, timestamp time.Time, file_pwd string, file_name string) error {
	// TO DO: Handle duplicates being added to the db? What happens if duplicates are added with the same timestamps and host, but different json_data
	snapshot := Snapshot{
		UUID:      uuid.New(),
		Host_IP:   host_ip,
		Timestamp: timestamp,
		File_PWD:  file_pwd,
		File_Name: file_name,
	}
	err := sr.db.WithContext(ctx).Create(snapshot).Error
	return err
}

func (sr *snapshotRepo) GetSnapshotByFileName(ctx context.Context, host_ip string, filename string) (Snapshot, error) {
	var snapshot Snapshot
	err := sr.db.WithContext(ctx).Where(
		"host_ip = ? AND File_PWD = ?",
		host_ip, filename,
	).First(&snapshot).Error
	if err != nil {
		return snapshot, err
	}
	return snapshot, nil
}

func (sr *snapshotRepo) GetSnapshotByTimeStamp(ctx context.Context, host_ip string, timestamp time.Time) (Snapshot, error) {
	var snapshot Snapshot
	err := sr.db.WithContext(ctx).Where(
		"host_ip = ? AND timestamp = ?",
		host_ip, timestamp,
	).First(&snapshot).Error
	if err != nil {
		return snapshot, err
	}
	return snapshot, nil
}

func (sr *snapshotRepo) GetAllHosts(ctx context.Context) ([]string, error) {
	var hosts []string
	err := sr.db.WithContext(ctx).Model(&Snapshot{}).Distinct("host_ip").Pluck("host_ip", &hosts).Error
	if err != nil {
		return []string{}, err
	}
	return hosts, nil
}

func (sr *snapshotRepo) ListAllHostSnapshots(ctx context.Context, host_ip string) ([]string, error) {
	var availSnapshots []Snapshot
	err := sr.db.WithContext(ctx).Where("host_ip = ?", host_ip).Find(&availSnapshots).Error
	if err != nil {
		return []string{}, err
	}
	timestamps := make([]string, 0, len(availSnapshots))
	for _, snapshot := range availSnapshots {
		timestamps = append(timestamps, snapshot.Timestamp.Format(time.RFC3339))
	}
	return timestamps, nil
}

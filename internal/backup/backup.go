package backup

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"github.com/SmoothWay/metrics/internal/service"
)

type BackupConfig struct {
	Interval int64
	FilePath string
	s        *service.Service
}

func New(interval int64, path string, serv *service.Service) (*BackupConfig, error) {

	return &BackupConfig{
		Interval: interval,
		FilePath: path,
		s:        serv,
	}, nil

}

var ErrRestoreFromFile = errors.New("error restoring from file")

func (b *BackupConfig) Backup(ctx context.Context) error {
	backupInterval := time.NewTicker(time.Duration(b.Interval) * time.Second)
	defer backupInterval.Stop()

	for {
		select {
		case <-backupInterval.C:
			if err := b.backupToFile(); err != nil {
				return err
			}
		case <-ctx.Done():
			return b.backupToFile()
		}
	}
}

func (b *BackupConfig) backupToFile() error {
	metrics := b.s.GetAll()

	if len(metrics) == 0 {
		return nil
	}

	logger.Log().Info("writing to file", zap.Int("num of metrics", len(metrics)))
	file, err := os.Create(b.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(metrics); err != nil {
		logger.Log().Error("Error by encode metrics to json", zap.Error(err))
		return err
	}
	return nil
}

func Restore(FilePath string) (*[]model.Metrics, error) {
	file, err := os.OpenFile(FilePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Println("error opening file", err)
		return nil, err
	}
	defer file.Close()

	var metrics []model.Metrics
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		log.Println("error by decode metrics from json", err)
		return &metrics, ErrRestoreFromFile
	}

	log.Println("restored metrics from file")
	return &metrics, nil
}

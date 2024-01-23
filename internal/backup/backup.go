package backup

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"github.com/SmoothWay/metrics/internal/repository/memstorage"
	"github.com/SmoothWay/metrics/internal/service"
	"go.uber.org/zap"
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

func (b *BackupConfig) Backup(ctx context.Context) error {
	backupInterval := time.NewTicker(time.Duration(b.Interval) * time.Second)
	for {
		metrics := b.s.GetAll()
		select {
		case <-backupInterval.C:
			err := b.saveTofile(metrics)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			err := b.saveTofile(metrics)
			return err
		}

	}
}

var ErrRestoreFromJSON = errors.New("error restoring from JSON")

func Restore(FilePath string) (*memstorage.MemStorage, error) {
	file, err := os.OpenFile(FilePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Log.Error("Error opening file", zap.Error(err))
		return nil, err
	}
	m := memstorage.New()
	if err := json.NewDecoder(file).Decode(m); err != nil {
		logger.Log.Error("Error by decode metrics from json", zap.Error(err))
		return m, ErrRestoreFromJSON
	}
	log.Printf("restored from file: %+v\n", m)
	return m, nil
}

func (b *BackupConfig) saveTofile(m []model.Metrics) error {

	logger.Log.Info("writing to file", zap.Any("bytes", m))
	file, err := os.OpenFile(b.FilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(m); err != nil {
		logger.Log.Error("Error by encode metrics to json", zap.Error(err))
		return err
	}
	return nil
}

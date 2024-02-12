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

		select {
		case <-backupInterval.C:
			metrics := b.s.GetAll()

			if len(metrics) == 0 {
				continue
			}
			err := b.saveTofile(metrics)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			metrics := b.s.GetAll()

			if len(metrics) == 0 {
				return nil
			}
			err := b.saveTofile(metrics)
			return err
		}

	}
}

var ErrRestoreFromFile = errors.New("error restoring from file")

func Restore(FilePath string) (*[]model.Metrics, error) {
	file, err := os.OpenFile(FilePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("error opening file", err)
		return nil, err
	}
	metrics := make([]model.Metrics, 0)

	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		log.Println("error by decode metrics from json", err)
		return &metrics, ErrRestoreFromFile
	}

	log.Println("restored metrics from file")
	return &metrics, nil
}

func (b *BackupConfig) saveTofile(m []model.Metrics) error {

	logger.Log().Info("writing to file", zap.Int("num of metrics", len(m)))
	file, err := os.OpenFile(b.FilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(m); err != nil {
		logger.Log().Error("Error by encode metrics to json", zap.Error(err))
		return err
	}
	return nil
}

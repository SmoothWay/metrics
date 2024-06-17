package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/backup"
	"github.com/SmoothWay/metrics/internal/config"
	"github.com/SmoothWay/metrics/internal/crypt"
	gserver "github.com/SmoothWay/metrics/internal/grpc/server"
	"github.com/SmoothWay/metrics/internal/handler"
	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"github.com/SmoothWay/metrics/internal/repository/memstorage"
	"github.com/SmoothWay/metrics/internal/repository/postgres"
	"github.com/SmoothWay/metrics/internal/service"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	cfg := config.NewServerConfig()
	var repo service.Repository
	var metrics *[]model.Metrics

	err := logger.Init(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Restore {
		metrics, err = backup.Restore(cfg.StoragePath)
		if err != nil {
			if errors.Is(backup.ErrRestoreFromFile, err) {
				log.Println("cant restore from json")
			} else {
				log.Println("unexpected err restoring from json", zap.Error(err))
			}
		}
	}
	if cfg.DSN != "" {
		repo, err = postgres.New(cfg.DSN)
		if err != nil {
			log.Fatal("error init postgres:", err)
		}
	} else {
		repo = memstorage.New(metrics)
	}
	serv := service.New(repo)

	cfg.B, err = backup.New(cfg.StoreInvterval, cfg.StoragePath, serv)
	if err != nil {
		log.Fatal("err creating backupper", zap.Error(err))
	}

	// config.H = handler.NewHandler(serv)

	logger.Log().Info("server version", zap.String("version", buildVersion), zap.String("build_date", buildDate), zap.String("build_commit", buildCommit))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if cfg.StoreInvterval > 0 {
		ticker := time.NewTicker(time.Duration(cfg.StoreInvterval))
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-ctx.Done():
					logger.Log().Info("Context cancelled. Stopping backup routine.")
					return
				case <-ticker.C:
					err = cfg.B.Backup(ctx)
					if err != nil {
						logger.Log().Error("Backup encountered error", zap.Error(err))
					}
				}
			}
		}()
	}

	var privateKey []byte
	if cfg.CryptKeyPath != "" {
		privateKey, err = crypt.ReadKeyFile(cfg.CryptKeyPath)
		if err != nil {
			logger.Log().Error("readKeyFile", zap.Error(err))
			return
		}
	}

	switch cfg.ServerType {
	case model.HTTPType:
		s := handler.NewServer(cfg.Host, handler.NewHandler(serv), cfg.Key, cfg.TrustedSubnet, privateKey)
		go func() {
			logger.Log().Info("Starting server on", zap.String("host", cfg.Host))
			if err := s.Run(); err != nil && err != http.ErrServerClosed {
				logger.Log().Fatal("Failed to start server", zap.Error(err))
			}
		}()
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()

		if err := s.Shutdown(shutdownCtx); err != nil {
			logger.Log().Error("Server shutdown failed", zap.Error(err))
		} else {
			logger.Log().Info("Server gracefully stopped")
		}
	case model.GRPCType:
		repo = memstorage.New(metrics)

		serv := service.New(repo)
		grpcServer := gserver.NewServer(gserver.Config{
			ServerAddr:    cfg.Host,
			Service:       serv,
			TrustedSubnet: handler.TrustedSubnetFromString(cfg.TrustedSubnet),
		})

		go grpcServer.Run(ctx)

		<-ctx.Done()

		if err = grpcServer.Shutdown(ctx); err != nil {
			logger.Log().Error("Server shutdown failed", zap.Error(err))
		} else {
			logger.Log().Info("Server gracefully stopped")
		}

	default:
		logger.Log().Error("run server", zap.String("error", "invalid server type"))
	}

}

package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/crypt"
	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
)

var counter int64

type Agent struct {
	PubKey  []byte
	Host    string
	Key     string
	Client  *http.Client
	Metrics []model.Metrics
	mu      sync.Mutex
}

// ReportAllMetricsAtOnes - sends all collected metrics in one single slice to jobs channel
func (a *Agent) ReportAllMetricsAtOnes(ctx context.Context, jobs chan<- []model.Metrics) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	jobs <- a.Metrics
}

// Worker - worker which sends request of single instance of metric to server
func (a *Agent) Worker(ctx context.Context, id int, jobs <-chan []model.Metrics, errs chan<- error) {
	for {
		select {
		case <-ctx.Done():
			logger.Log().Info("worker done", zap.Int("id", id))
			return
		case metrics, ok := <-jobs:
			if !ok {
				logger.Log().Info("worker done", zap.Int("id", id))
				return
			}
			logger.Log().Info("worker", zap.Int("started id", id))
			for _, metric := range metrics {
				err := a.sendRequest(ctx, metric)
				if err != nil {
					errs <- err
				}
			}
		}
	}
}

// Retry - retry mechanism for sedning request to server again if request failed
func (a *Agent) Retry(ctx context.Context, numRetry int, jobs chan []model.Metrics, fn func(chan []model.Metrics)) {
	fn(jobs)

	for i := 1; i <= numRetry; i++ {
		logger.Log().Info("Retry", zap.Int("Retrying...", i))

		retryTicker := time.NewTicker(time.Duration(i+2) * time.Second)

		select {
		case <-retryTicker.C:
			fn(jobs)
		case <-ctx.Done():
			return
		}
	}

}

func compressData(data []byte) (io.Reader, error) {
	b := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(b, gzip.BestSpeed)
	if err != nil {
		logger.Log().Error("error init gzip writer", zap.Error(err))
		return nil, err
	}
	_, err = w.Write(data)
	if err != nil {
		logger.Log().Error("error compressing data", zap.Error(err))
		return nil, err
	}
	err = w.Close()
	w.Reset(b)
	if err != nil {
		logger.Log().Error("error closing writer", zap.Error(err))
		return nil, err
	}
	return b, nil
}

func (a *Agent) sendRequest(ctx context.Context, m model.Metrics) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	jsonMetric, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if len(a.PubKey) > 0 {
		jsonMetric, err = crypt.Encrypt(jsonMetric, a.PubKey)
		if err != nil {
			return err
		}
	}

	cJSONMetric, err := compressData(jsonMetric)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s/update/", a.Host)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, cJSONMetric)
	if err != nil {
		return err
	}
	logger.Log().Info("sent requset")
	if a.Key != "" {

		h := hmac.New(sha256.New, []byte(a.Key))

		h.Write(jsonMetric)
		metricsHash := h.Sum(nil)

		hashString := hex.EncodeToString(metricsHash)

		req.Header.Add("HashSHA256", string(hashString))
	}

	defer req.Body.Close()
	ip, err := getIP()
	if err != nil {
		logger.Log().Warn("cant get ip", zap.String("error", err.Error()))
	}
	req.Header.Set("X-REAL-IP", ip.String())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	res, err := a.Client.Do(req)
	if err != nil {
		return err
	}

	res.Body.Close()
	return nil
}

// ReportMetrics - send slice of metrics to server with compression
func (a *Agent) ReportMetrics(ctx context.Context) error {

	var wg sync.WaitGroup
	errChan := make(chan error, len(a.Metrics))

	for _, m := range a.Metrics {
		m := m
		wg.Add(1)

		go func(m model.Metrics) {
			defer wg.Done()

			jsonMetric, err := json.Marshal(m)
			if err != nil {
				errChan <- err
				return
			}
			cJSONMetric, err := compressData(jsonMetric)
			if err != nil {
				errChan <- err
				return
			}

			endpoint := fmt.Sprintf("http://%s/update/", a.Host)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, cJSONMetric)
			if err != nil {
				errChan <- err
				return
			}
			defer req.Body.Close()
			ip, err := getIP()
			if err != nil {
				logger.Log().Warn("cant get ip", zap.String("error", err.Error()))
			}
			log.Println(ip.String())
			req.Header.Set("X-REAL-IP", ip.String())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding", "gzip")

			res, err := a.Client.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			res.Body.Close()
		}(m)

	}
	go func() {
		wg.Wait()
		close(errChan)
	}()

	var errors []error

	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors: %v", len(errors), errors)
	}

	return nil

}

func getIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var l []net.IP
	var p []net.IP

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				l = append(l, ipnet.IP)
				continue
			}

			if ipnet.IP.IsPrivate() && ipnet.IP.To4() != nil {
				p = append(p, ipnet.IP)
				continue
			}

			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			}
		}
	}

	if len(p) > 0 {
		return p[0], nil
	}

	if len(l) > 0 {
		return l[0], nil
	}

	return nil, nil
}

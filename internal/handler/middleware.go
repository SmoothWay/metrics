package handler

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/SmoothWay/metrics/internal/crypt"
	"github.com/SmoothWay/metrics/internal/logger"
)

type responseData struct {
	status int
	size   int
}

type Middleware struct {
	HashSecretKey string
}

func NewMiddleware(hash string) *Middleware {
	return &Middleware{
		HashSecretKey: hash,
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (mw *Middleware) Decrypt(privateKey []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if len(privateKey) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log().Error("Decrypt", zap.Error(err))
				writeJSON(w, http.StatusBadRequest, err.Error())
				return
			}

			body, err = crypt.Decrypt(body, privateKey)
			if err != nil {
				logger.Log().Error("Decrypt", zap.Error(err))
				writeJSON(w, http.StatusBadRequest, "Failed to decrypt request data")
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func (mw *Middleware) requestLogger(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		duration := time.Since(start)
		next.ServeHTTP(&lw, r)
		logger.Log().Info("incomming request",
			zap.String("uri", uri),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("code", lw.responseData.status),
			zap.Int("size", lw.responseData.size),
		)

	})
}

func (mw *Middleware) decompresser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gz
			defer gz.Close()
		}

		acceptedEncodings := r.Header.Get("Accept-Encoding")
		canCompressResponse := strings.Contains(acceptedEncodings, "gzip")

		if canCompressResponse {
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()
			w = gzipResponseWriter{Writer: gzipWriter, ResponseWriter: w}
			w.Header().Set("Content-Encoding", "gzip")
		}

		next.ServeHTTP(w, r)
	})
}

func (mw *Middleware) checkHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hash := r.Header.Get("HashSHA256"); hash != "" {
			h := hmac.New(sha256.New, []byte(mw.HashSecretKey))

			bodyReader := io.TeeReader(r.Body, h)

			r.Body = io.NopCloser(bodyReader)

			io.Copy(io.Discard, r.Body)

			r.Body = io.NopCloser(bytes.NewReader([]byte{}))

			metricsHash := h.Sum(nil)
			strHash := hex.EncodeToString(metricsHash)

			if strHash != hash {
				badRequestResponse(w, r, errors.New("hash mismatch"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (mw *Middleware) TrustedSubnet(subnet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if subnet == nil {
				next.ServeHTTP(w, r)
				return
			}

			contentType := r.Header.Get("Content-Type")

			remoteAddr := r.Header.Get("X-REAL-IP")

			if remoteAddr == "" {
				ra := strings.Split(r.RemoteAddr, ":")
				if len(ra) >= 2 {
					remoteAddr = ra[0]
				}
			}

			ip := net.ParseIP(remoteAddr)

			if ip == nil || !subnet.Contains(ip) {
				if strings.Contains(contentType, "application/json") {
					log.Println(ip)
					writeJSON(w, http.StatusForbidden, "request from this ip-address was rejected")
				} else {
					http.Error(w, "request from this ip-address was rejected", http.StatusForbidden)
				}
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

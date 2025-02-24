package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/mediaprodcast/storage/pkg/backend/server/utils"
	"github.com/mediaprodcast/storage/pkg/storage/provider"
)

type ActiveUpload struct {
	mu        sync.RWMutex
	buffer    []byte
	eof       bool
	header    http.Header
	createdAt time.Time
	maxAge    int64
}

type StreamingHandler struct {
	activeUploads map[string]*ActiveUpload
	uploadsLock   sync.RWMutex
	stopChan      chan struct{}
}

func NewStreamingHandler() *StreamingHandler {
	h := &StreamingHandler{
		activeUploads: make(map[string]*ActiveUpload),
		stopChan:      make(chan struct{}),
	}
	go h.cleanupActiveUploads()
	return h
}

func (h *StreamingHandler) HandleChunkedUpload(
	ctx context.Context,
	storageBackend provider.Storage,
	w http.ResponseWriter,
	r *http.Request,
	path string,
) {
	au := &ActiveUpload{
		header:    r.Header.Clone(),
		createdAt: time.Now(),
		maxAge:    getMaxAgeOr(r.Header.Get("Cache-Control"), -1),
	}

	h.uploadsLock.Lock()
	h.activeUploads[path] = au
	h.uploadsLock.Unlock()

	defer h.cleanupUpload(path)

	r.Body = http.MaxBytesReader(w, r.Body, utils.MaxUploadSize)
	defer r.Body.Close()

	buf := make([]byte, 32*1024) // 32KB chunks
	for {
		n, err := r.Body.Read(buf)
		if n > 0 {
			au.mu.Lock()
			au.buffer = append(au.buffer, buf[:n]...)
			au.mu.Unlock()
		}

		if err != nil {
			if err != io.EOF {
				utils.WriteError(w, "Upload failed", http.StatusBadRequest, err)
				return
			}
			break
		}
	}

	au.mu.RLock()
	defer au.mu.RUnlock()
	if err := storageBackend.Save(context.Background(), bytes.NewReader(au.buffer), path); err != nil {
		utils.WriteError(w, "Final save failed", http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *StreamingHandler) GetActiveUpload(path string) (*ActiveUpload, bool) {
	h.uploadsLock.RLock()
	defer h.uploadsLock.RUnlock()
	au, exists := h.activeUploads[path]
	return au, exists
}

func (h *StreamingHandler) cleanupUpload(path string) {
	h.uploadsLock.Lock()
	delete(h.activeUploads, path)
	h.uploadsLock.Unlock()
}

func (h *StreamingHandler) cleanupActiveUploads() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.uploadsLock.Lock()
			now := time.Now()
			// Check for expired files
			for path, au := range h.activeUploads {
				au.mu.RLock()
				expirationTime := au.createdAt.Add(time.Second * time.Duration(au.maxAge))
				if expirationTime.Before(now) {
					delete(h.activeUploads, path)
				}
				au.mu.RUnlock()
			}
			h.uploadsLock.Unlock()
		case <-h.stopChan:
			return
		}
	}
}

func (h *StreamingHandler) Shutdown() {
	close(h.stopChan)
}

func getMaxAgeOr(s string, def int64) int64 {
	ret := def
	r := regexp.MustCompile(`max-age=(?P<maxage>\d*)`)
	match := r.FindStringSubmatch(s)
	for i, name := range r.SubexpNames() {
		if i > 0 && i <= len(match) {
			if name == "maxage" {
				valInt, err := strconv.ParseInt(match[i], 10, 64)
				if err == nil {
					ret = valInt
					break
				}
			}
		}
	}
	return ret
}

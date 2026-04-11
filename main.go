package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var publicDir = "./public"

const reloadScript = `<script>
(function(){
  var es = new EventSource("/_reload");
  es.onmessage = function(e){ if (e.data === "reload") location.reload(); };
})();
</script>`

func main() {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "public")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			publicDir = candidate
		}
	}

	go watchLoop()

	http.HandleFunc("/_reload", sseHandler)
	http.HandleFunc("/", serveHandler)

	addr := ":8080"
	log.Printf("serving %s on http://localhost%s", publicDir, addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func serveHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	if strings.HasSuffix(urlPath, "/") {
		urlPath += "index.html"
	}
	full := filepath.Join(publicDir, filepath.FromSlash(urlPath))

	absRoot, _ := filepath.Abs(publicDir)
	absFull, _ := filepath.Abs(full)
	if !strings.HasPrefix(absFull, absRoot) {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(full)
	if err != nil || info.IsDir() || !strings.HasSuffix(urlPath, ".html") {
		http.FileServer(http.Dir(publicDir)).ServeHTTP(w, r)
		return
	}

	data, err := os.ReadFile(full)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if idx := bytes.LastIndex(bytes.ToLower(data), []byte("</body>")); idx >= 0 {
		data = append(append(data[:idx:idx], []byte(reloadScript)...), data[idx:]...)
	} else {
		data = append(data, []byte(reloadScript)...)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

var (
	mu        sync.Mutex
	listeners = map[chan struct{}]bool{}
	lastHash  [32]byte
	zeroHash  [32]byte
)

func watchLoop() {
	for {
		h := snapshot()
		mu.Lock()
		if h != lastHash && lastHash != zeroHash {
			for c := range listeners {
				select {
				case c <- struct{}{}:
				default:
				}
			}
		}
		lastHash = h
		mu.Unlock()
		time.Sleep(500 * time.Millisecond)
	}
}

func snapshot() [32]byte {
	h := sha256.New()
	filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		fmt.Fprintf(h, "%s|%d|%d\n", path, info.Size(), info.ModTime().UnixNano())
		return nil
	})
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan struct{}, 1)
	mu.Lock()
	listeners[ch] = true
	mu.Unlock()
	defer func() {
		mu.Lock()
		delete(listeners, ch)
		mu.Unlock()
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			io.WriteString(w, "data: reload\n\n")
			flusher.Flush()
		}
	}
}

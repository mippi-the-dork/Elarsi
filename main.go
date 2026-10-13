package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//go:embed web/*
var webFiles embed.FS

type sourceRegistry struct {
	sync.Mutex
	paths map[string]string
}

func validKind(k string) bool { return k == "audio" || k == "lyrics" || k == "project" }
func pickerFilter(k string) string {
	switch k {
	case "audio":
		return "Audio files (*.mp3;*.wav;*.flac;*.ogg;*.m4a)|*.mp3;*.wav;*.flac;*.ogg;*.m4a|All files (*.*)|*.*"
	case "lyrics":
		return "Lyrics (*.txt;*.lrc)|*.txt;*.lrc|All files (*.*)|*.*"
	default:
		return "Elarsi projects (*.elarsi;*.lyricsync;*.json)|*.elarsi;*.lyricsync;*.json|All files (*.*)|*.*"
	}
}
func selectWindowsFile(k string) (string, error) {
	// The filter comes only from the allowlisted kind, not from browser input.
	ps := `Add-Type -AssemblyName System.Windows.Forms; $d=New-Object System.Windows.Forms.OpenFileDialog; $d.Filter='` + pickerFilter(k) + `'; $d.CheckFileExists=$true; if ($d.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { [Console]::Out.Write($d.FileName) }`
	out, err := exec.Command("powershell.exe", "-NoProfile", "-STA", "-Command", ps).Output()
	return strings.TrimSpace(string(out)), err
}
func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	host := listener.Addr().String()
	var lastPing atomic.Int64
	lastPing.Store(time.Now().Unix())
	registry := &sourceRegistry{paths: make(map[string]string)}
	mux := http.NewServeMux()
	site, _ := fs.Sub(webFiles, "web")
	mux.Handle("/", http.FileServer(http.FS(site)))
	mux.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) { lastPing.Store(time.Now().Unix()); w.WriteHeader(204) })
	mux.HandleFunc("/native/open", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		k := r.URL.Query().Get("kind")
		if !validKind(k) {
			http.Error(w, "Invalid source type", 400)
			return
		}
		if runtime.GOOS != "windows" {
			http.Error(w, "Windows only", 501)
			return
		}
		path, err := selectWindowsFile(k)
		if err != nil {
			http.Error(w, "File picker failed", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if path == "" {
			json.NewEncoder(w).Encode(map[string]any{"cancelled": true})
			return
		}
		registry.Lock()
		registry.paths[k] = path
		registry.Unlock()
		json.NewEncoder(w).Encode(map[string]any{"name": filepath.Base(path), "path": path})
	})
	mux.HandleFunc("/native/file", func(w http.ResponseWriter, r *http.Request) {
		k := r.URL.Query().Get("kind")
		if !validKind(k) {
			http.Error(w, "Invalid source type", 400)
			return
		}
		registry.Lock()
		path := registry.paths[k]
		registry.Unlock()
		if path == "" {
			http.Error(w, "No file selected", 404)
			return
		}
		f, err := os.Open(path)
		if err != nil {
			http.Error(w, "Cannot open selected file", 404)
			return
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			http.Error(w, "Invalid file", 404)
			return
		}
		w.Header().Set("Content-Disposition", `inline; filename="`+strings.ReplaceAll(filepath.Base(path), `"`, ``)+`"`)
		http.ServeContent(w, r, filepath.Base(path), stat.ModTime(), f)
	})
	mux.HandleFunc("/native/reveal", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		k := r.URL.Query().Get("kind")
		if !validKind(k) {
			http.Error(w, "Invalid source type", 400)
			return
		}
		registry.Lock()
		path := registry.paths[k]
		registry.Unlock()
		if path == "" {
			http.Error(w, "Source location unavailable; use Browse to import it", 404)
			return
		}
		if _, err := os.Stat(path); err != nil {
			http.Error(w, "Original file not found", 404)
			return
		}
		if err := exec.Command("explorer.exe", "/select,", path).Start(); err != nil {
			http.Error(w, "Could not open Explorer", 500)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST required", 405)
			return
		}
		w.WriteHeader(204)
		go func() { time.Sleep(250 * time.Millisecond); os.Exit(0) }()
	})
	// Prevent websites on other origins from triggering native file dialogs or Explorer actions.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != host {
			http.Error(w, "Invalid host", 403)
			return
		}
		if o := r.Header.Get("Origin"); o != "" && o != "http://"+host {
			http.Error(w, "Invalid origin", 403)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/native/") && r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			http.Error(w, "Cross-site requests blocked", 403)
			return
		}
		mux.ServeHTTP(w, r)
	})
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	go server.Serve(listener)
	url := fmt.Sprintf("http://%s", host)
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	default:
		_ = exec.Command("xdg-open", url).Start()
	}
	for range time.Tick(10 * time.Second) {
		if time.Now().Unix()-lastPing.Load() > 55 {
			_ = server.Close()
			return
		}
	}
}

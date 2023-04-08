package main

import (
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/config"
	"io"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntireAppInsecure(t *testing.T) {
	time.AfterFunc(time.Millisecond*100, func() {
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	})

	stdOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	_ = w.Close()
	result, _ := io.ReadAll(r)
	output := string(result)
	os.Stdout = stdOut

	assert.Contains(t, output, "Starting insecure server")
}

func TestEntireAppSecured(t *testing.T) {
	config.Settings.EnableHTTPS = true
	time.AfterFunc(time.Millisecond*100, func() {
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	})

	stdOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	_ = w.Close()
	result, _ := io.ReadAll(r)
	output := string(result)
	os.Stdout = stdOut

	assert.Contains(t, output, "Starting HTTPS server")
}

package mergereader_test

import (
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/leafbridge/leafbridge-deploy/internal/mergereader"
)

func TestReader(t *testing.T) {
	r1, w1, err1 := os.Pipe()
	r2, w2, err2 := os.Pipe()
	if err := errors.Join(err1, err2); err != nil {
		t.Fatal(err)
	}

	go func() {
		for range 10 {
			w1.WriteString("Hello from pipe 1\n")
			time.Sleep(10 * time.Millisecond)
		}
		w1.Close()
	}()

	go func() {
		for range 10 {
			w2.WriteString("Hello from pipe 2\n")
			time.Sleep(10 * time.Millisecond)
		}
		w2.Close()
	}()

	merged := mergereader.New(r1, r2)
	data, err := io.ReadAll(merged)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))
}

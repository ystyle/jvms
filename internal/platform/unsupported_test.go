//go:build !windows

package platform

import (
	"testing"

	appcfg "github.com/ystyle/jvms/internal/config"
)

func TestUnsupportedProviderStreamEndsWithError(t *testing.T) {
	events := NewProvider(appcfg.NewConfig()).StreamAvailable()
	event, ok := <-events
	if !ok || event.Err == nil || !event.Done {
		t.Fatalf("expected terminal platform error, got %+v, open=%t", event, ok)
	}
	if _, ok := <-events; ok {
		t.Fatal("stream must close after the platform error")
	}
}

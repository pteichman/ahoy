package ahoy

import (
	"log"
	"os"
	"testing"

	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/sdk/trace"
)

func newTestEnv(t *testing.T) (*Env, func()) {
	t.Helper()

	exporter, err := stdout.NewExporter(stdout.WithPrettyPrint())
	if err != nil {
		t.Fatalf("stdout.NewExporter: %s", err)
	}

	config := trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	}

	provider, err := trace.NewProvider(trace.WithConfig(config), trace.WithSyncer(exporter))
	if err != nil {
		t.Fatalf("trace.NewProvider: %s", err)
	}

	env := &Env{
		PublicHost: "example.org",
		PublicURL:  "https://example.org",
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
		Tracer:     provider.Tracer("ahoy_test"),
	}

	cleanup := func() {
	}

	return env, cleanup
}

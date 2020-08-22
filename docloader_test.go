package ahoy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDocLoader_LoadDocument(t *testing.T) {
	tests := []struct {
		name string

		contentType string
		body        string

		want string
	}{
		{
			"0002-in.json",
			"application/ld+json",
			`{
  "@context": {
    "@vocab": "http://example/vocab#"
  },
  "@id": "",
  "term": "object"
}`,
			`{"@context":{"@vocab":"http://example/vocab#"},"@id":"","term":"object"}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.Write([]byte(tt.body))
			}

			srv := httptest.NewServer(http.HandlerFunc(handler))
			defer srv.Close()

			dl := &DefaultDocumentLoader{
				httpClient: &http.Client{},
			}

			remoteDoc, err := dl.LoadDocument(srv.URL)
			if err != nil {
				t.Errorf("LoadDocument -> %s", err)
			}

			if remoteDoc.DocumentURL != srv.URL {
				t.Errorf("DocumentURL=%s, want %s", remoteDoc.DocumentURL, srv.URL)
			}

			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(remoteDoc.Document); err != nil {
				t.Errorf("Encode() -> %s", err)
			}

			doc := buf.String()
			if diff := cmp.Diff(tt.want, doc); diff != "" {
				t.Errorf("Doc mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

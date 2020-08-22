package ahoy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/piprate/json-gold/ld"
)

const (
	// An HTTP Accept header that prefers JSONLD.
	acceptHeader = "application/ld+json, application/json;q=0.9, application/javascript;q=0.5, text/javascript;q=0.5, text/plain;q=0.2, */*;q=0.1"

	ApplicationJSONLDType = "application/ld+json"

	// JSON-LD link header rel
	linkHeaderRel = "http://www.w3.org/ns/json-ld#context"
)

var rApplicationJSON = regexp.MustCompile(`^application/(\w*\+)?json$`)

// DefaultDocumentLoader is a standard implementation of DocumentLoader
// which can retrieve documents via HTTP.
type DefaultDocumentLoader struct {
	httpClient *http.Client
}

// NewDefaultDocumentLoader creates a new instance of DefaultDocumentLoader
func NewDefaultDocumentLoader(httpClient *http.Client) *DefaultDocumentLoader {
	rval := &DefaultDocumentLoader{httpClient: httpClient}

	if rval.httpClient == nil {
		rval.httpClient = http.DefaultClient
	}
	return rval
}

// DocumentFromReader returns a document containing the contents of the JSON resource,
// streamed from the given Reader.
func DocumentFromReader(r io.Reader) (interface{}, error) {
	var document interface{}
	dec := json.NewDecoder(r)

	// If dec.UseNumber() were invoked here, all numbers would be decoded as json.Number.
	// json-gold supports both the default and json.Number options.

	if err := dec.Decode(&document); err != nil {
		return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
	}
	return document, nil
}

// LoadDocument returns a ld.RemoteDocument containing the contents of the JSON resource
// from the given URL.
func (dl *DefaultDocumentLoader) LoadDocument(u string) (*ld.RemoteDocument, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, fmt.Sprintf("error parsing URL: %s", u))
	}

	remoteDoc := &ld.RemoteDocument{}

	protocol := parsedURL.Scheme
	if protocol != "http" && protocol != "https" {
		// Can't use the HTTP client for those!
		remoteDoc.DocumentURL = u
		var file *os.File
		file, err = os.Open(u)
		if err != nil {
			return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
		}
		defer file.Close()

		remoteDoc.Document, err = DocumentFromReader(file)
		if err != nil {
			return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
		}
	} else {

		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
		}
		// We prefer application/ld+json, but fallback to application/json
		// or whatever is available
		req.Header.Add("Accept", acceptHeader)

		res, err := dl.httpClient.Do(req)
		if err != nil {
			return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed,
				fmt.Sprintf("Bad response status code: %d", res.StatusCode))
		}

		remoteDoc.DocumentURL = res.Request.URL.String()

		contentType := res.Header.Get("Content-Type")
		linkHeader := res.Header.Get("Link")

		if len(linkHeader) > 0 {
			parsedLinkHeader := ld.ParseLinkHeader(linkHeader)
			contextLink := parsedLinkHeader[linkHeaderRel]
			if contextLink != nil && contentType != ApplicationJSONLDType &&
				(contentType == "application/json" || rApplicationJSON.MatchString(contentType)) {

				if len(contextLink) > 1 {
					return nil, ld.NewJsonLdError(ld.MultipleContextLinkHeaders, nil)
				} else if len(contextLink) == 1 {
					remoteDoc.ContextURL = contextLink[0]["target"]
				}
			}

			// If content-type is not application/ld+json, nor any other +json
			// and a link with rel=alternate and type='application/ld+json' is found,
			// use that instead
			alternateLink := parsedLinkHeader["alternate"]
			if alternateLink != nil &&
				alternateLink[0]["type"] == ApplicationJSONLDType &&
				!rApplicationJSON.MatchString(contentType) {

				finalURL := ld.Resolve(u, alternateLink[0]["target"])
				return dl.LoadDocument(finalURL)
			}
		}

		remoteDoc.Document, err = DocumentFromReader(res.Body)
		if err != nil {
			return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
		}
	}
	return remoteDoc, nil
}

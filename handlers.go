package ahoy

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/piprate/json-gold/ld"
	"go.opentelemetry.io/otel/api/trace"
	"google.golang.org/grpc/codes"
)

type Env struct {
	PublicHost string
	PublicURL  string

	Logger *log.Logger
	Tracer trace.Tracer
}

func handleUsers(env *Env) httprouter.Handle {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	context := []interface{}{
		"https://www.w3.org/ns/activitystreams",
		"https://w3id.org/security/v1",
	}

	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx, span := env.Tracer.Start(r.Context(), "ahoy.handleUsers")
		defer span.End()

		doc := map[string]interface{}{
			"id":                r.URL.String(),
			"type":              "Person",
			"preferredUsername": "username",
			"inbox":             env.PublicURL + "/inbox/username",
		}

		doc2, err := proc.Compact(doc, context, options)
		if err != nil {
			span.RecordError(ctx, err, trace.WithErrorStatus(codes.Internal))

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(doc2)
		if err != nil {
			span.RecordError(ctx, err, trace.WithErrorStatus(codes.Internal))

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func handleActor(env *Env) httprouter.Handle {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	//	documentContext2 := map[string]interface{}{
	//		"@context": "https://www.w3.org/ns/activitystreams",
	//		"@context": "https://w3id.org/security/v1",
	//	}

	// documentContext := []interface{}{
	// 	"https://www.w3.org/ns/activitystreams",
	// 	"https://w3id.org/security/v1",
	// }

	documentContext := map[string]interface{}{
		"as":  "https://www.w3.org/ns/activitystreams",
		"sec": "https://w3id.org/security/v1",
	}

	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		actor := `{
	"@context": [
		"https://www.w3.org/ns/activitystreams",
		"https://w3id.org/security/v1"
	],

	"id": "https://my-example.com/actor",
	"type": "Person",
	"preferredUsername": "alice",
	"inbox": "https://my-example.com/inbox",

	"publicKey": {
		"id": "https://my-example.com/actor#main-key",
		"owner": "https://my-example.com/actor",
		"publicKeyPem": "-----BEGIN PUBLIC KEY-----...-----END PUBLIC KEY-----"
	}
}`

		doc, err := ld.DocumentFromReader(bytes.NewReader([]byte(actor)))
		if err != nil {
			env.Logger.Println("aww", err)
			return
		}

		doc2, err := proc.Compact(doc, documentContext, options)
		if err != nil {
			env.Logger.Println("Error when compacting JSON-LD document:", err)
			return
		}

		json.NewEncoder(os.Stderr).Encode(doc)
		json.NewEncoder(os.Stderr).Encode(doc2)

		env.Logger.Printf("%+v", doc)
		env.Logger.Printf("%+v", doc2)

		env.Logger.Println("returning actor")

		// [map[@id:https://my-example.com/actor @type:[https://www.w3.org/ns/activitystreams#Person] http://www.w3.org/ns/ldp#inbox:[map[@id:https://my-example.com/inbox]] https://w3id.org/security#publicKey:[map[@id:https://my-example.com/actor#main-key https://w3id.org/security#owner:[map[@id:https://my-example.com/actor]] https://w3id.org/security#publicKeyPem:[map[@value:-----BEGIN PUBLIC KEY-----...-----END PUBLIC KEY-----]]]] https://www.w3.org/ns/activitystreams#preferredUsername:[map[@value:alice]]]]

		w.Write([]byte(actor))
	}
}

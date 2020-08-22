package ahoy

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/otel/api/trace"
	"google.golang.org/grpc/codes"
)

type WebfingerResponse struct {
	Subject string          `json:"subject"`
	Links   []WebfingerLink `json:"links"`
}

type WebfingerLink struct {
	Rel  string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}

func handleWebfinger(env *Env) httprouter.Handle {
	acctRe := regexp.MustCompile("acct:([a-z]+)@" + regexp.QuoteMeta(env.PublicHost))

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx, span := env.Tracer.Start(r.Context(), "ahoy.handleWebfinger")
		defer span.End()

		query := r.URL.Query()
		resourceParam := query.Get("resource")

		m := acctRe.FindStringSubmatch(resourceParam)
		if len(m) < 2 {
			span.RecordError(ctx, errors.New("not found"), trace.WithErrorStatus(codes.NotFound))
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		subject := m[0]
		username := m[1]

		resp := &WebfingerResponse{
			Subject: subject,
			Links: []WebfingerLink{
				{
					Rel:  "self",
					Type: "application/activity+json",
					Href: env.PublicURL + "/users/" + username,
				},
			},
		}

		w.Header().Set("Content-Type", "application/jrd+json")

		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			span.RecordError(ctx, err, trace.WithErrorStatus(codes.Internal))

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}

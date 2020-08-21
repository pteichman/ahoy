package ahoy

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/julienschmidt/httprouter"
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
		query := r.URL.Query()
		resourceParam := query.Get("resource")

		m := acctRe.FindStringSubmatch(resourceParam)
		if len(m) < 2 {
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
			env.Logger.Println("encoding jrd+json", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
}

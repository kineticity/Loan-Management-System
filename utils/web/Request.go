package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

func UnMarshalJSON(request *http.Request, out interface{}) error {
	//validation
	return json.NewDecoder(request.Body).Decode(out)

}

type Parser struct {
	Body   io.ReadCloser
	Form   url.Values
	Params map[string]string
}

func NewParser(request *http.Request) *Parser {
	return &Parser{
		Body:   request.Body,
		Form:   request.Form,
		Params: mux.Vars(request),
	}

}

// RespondWithJSON sends a JSON response to the client
func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// RespondWithError sends an error response to the client
func RespondWithError(w http.ResponseWriter, status int, message string) {
	RespondWithJSON(w, status, map[string]string{"error": message})
}

// GetUserIDFromContext extracts userID from the request context
func GetUserIDFromContext(r *http.Request) (uint, error) {
	userID, ok := r.Context().Value("user_id").(uint)
	fmt.Println(r.Context().Value("claims"))
	if !ok {
		return 0, errors.New("user ID not found in context")
	}
	return userID, nil
}

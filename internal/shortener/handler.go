package shortener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/go-playground/validator/v10"
	e "github.com/jasoncheung94/url-shortener/internal/errors"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	v "github.com/jasoncheung94/url-shortener/internal/validator"
)

// Handler represents the methods for handling CRUD.
type Handler struct {
	service Service
}

// NewHandler returns instance of Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes setup routes for shortener.
func (h *Handler) Routes(mux *http.ServeMux) {
	// GET
	mux.HandleFunc("/", HomeHandler)
	mux.HandleFunc("GET /favicon.ico", FaviconHandler)
	mux.HandleFunc("GET /{shorturl}", h.RedirectURL)
	mux.HandleFunc("GET /preview/{shorturl}", h.PreviewURL)

	// POST
	mux.HandleFunc("POST /shorten", h.ShortenURL)
}

// HomeHandler serves the HTML page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { // handle path that isn't registered!
		e.NewErrorResponse(http.StatusNotFound, "404 page not found", "404 page not found")
		return
	}

	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}
}

// FaviconHandler is a handler that returns the favicon.ico.
func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/assets/favicon.ico")

}

// ShortenURL creates a short URL
// @Summary Shortens a URL
// @Description Accepts a long URL, a custom alias, and an optional expiration date, and returns a shortened version
// @Tags URL Shortener
// @Accept json
// @Produce json
// @Param requestBody body model.URL true "Request body containing URL, custom alias, and expiration date"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /shorten [post]
func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestData model.URL
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		e.WriteJSONError(w, http.StatusBadRequest, e.NewErrorResponse(http.StatusBadRequest, "JSON error", err.Error()))
		return
	}
	r.Body.Close()

	// Validate request fields
	if err := v.Validate.Struct(requestData); err != nil {
		var errs []e.Error
		for _, fieldErr := range err.(validator.ValidationErrors) {
			errs = append(errs, e.Error{
				Status: http.StatusBadRequest,
				Title:  fmt.Sprintf("invalid field: %s", fieldErr.Field()),
				Detail: fmt.Sprintf("validation failed on '%s' tag", fieldErr.Tag()),
			})
		}
		e.WriteJSONError(w, http.StatusBadRequest, e.ErrorResponse{Errors: errs})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	data := &model.URL{
		OriginalURL:    requestData.OriginalURL,
		CustomURL:      requestData.CustomURL,
		ExpirationDate: requestData.ExpirationDate,
	}

	shortKey, err := h.service.SaveURL(ctx, data)
	switch {
	case errors.Is(err, e.ConflictError{}):
		e.WriteJSONError(w, http.StatusConflict, e.NewErrorResponse(http.StatusConflict, "conflict", err.Error()))
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response
	shortURL := "http://localhost:8080/" + shortKey
	data.ShortURL = shortURL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Encode the data as JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, fallback to a default error response
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// RedirectURL redirects a shortened URL to the original
// @Summary Redirects to the original URL
// @Description Finds the original URL from the shortened key and redirects
// @Tags URL Shortener
// @Param shorturl path string true "Shortened URL key"
// @Success 302
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /{shorturl} [get]
func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shorturl")
	if shortURL == "" || !isValidShortURL(shortURL) {
		e.WriteJSONError(w, http.StatusBadRequest, e.ErrorResponse{
			Errors: []e.Error{
				{
					Status: http.StatusBadRequest,
					Title:  "URL is not valid",
					Detail: "URL should be formatted correctly",
				},
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := h.service.GetURL(ctx, shortURL)
	switch {
	case errors.As(err, &e.NotFoundError{}), errors.Is(err, e.NotFoundError{}): // example of both.
		e.WriteJSONError(w, http.StatusNotFound, e.NewErrorResponse(http.StatusNotFound, "url not found", err.Error()))
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, data.OriginalURL, http.StatusFound)
}

// PreviewURL retrieves the original URL and related data for a given short URL.
// @Summary Preview a short URL
// @Description Returns information about a short URL, such as the original URL and metadata.
// @Tags URL Shortener
// @Accept json
// @Produce json
// @Param shorturl path string true "Short URL code"
// @Success 200 {object} model.URL
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {string} string
// @Router /preview/{shorturl} [get]
func (h *Handler) PreviewURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shortURL := r.PathValue("shorturl")
	if shortURL == "" || !isValidShortURL(shortURL) {
		e.WriteJSONError(w, http.StatusBadRequest, e.ErrorResponse{
			Errors: []e.Error{
				{
					Status: http.StatusBadRequest,
					Title:  "URL is not valid",
					Detail: "URL should be formatted correctly",
				},
			},
		})
		return
	}

	data, err := h.service.GetURL(ctx, shortURL)
	switch {
	case errors.As(err, &e.NotFoundError{}), errors.Is(err, e.NotFoundError{}): // example of both.
		e.WriteJSONError(w, http.StatusNotFound, e.NewErrorResponse(http.StatusNotFound, "url not found", err.Error()))
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lastModified := data.CreatedAt.UTC().Format(http.TimeFormat)
	w.Header().Set("Last-Modified", lastModified)           // App doesn't allow edit. Use created at date for http cache.
	w.Header().Set("Cache-Control", "public, max-age=3600") // http cache for 1 hour
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Encode the data as JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, fallback to a default error response
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// DumpStruct dumps data in a readable format.
// func DumpStruct(data any) {
// 	jsonData, err := json.MarshalIndent(data, "", "  ")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	fmt.Println(string(jsonData))
// }

package save

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"url-minimizer/internal/storage"
	response "url-minimizer/pkg/api/response"
	"url-minimizer/pkg/logger/sl"
	"url-minimizer/pkg/utils/random"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	Url   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias"`
}

type UrlSaver interface {
	Save(url string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver UrlSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var request Request
		err := render.DecodeJSON(r.Body, &request)
		if errors.Is(err, io.EOF) {
			log.Error("Request body is empty")
			responseError(w, r, "Empty request")
			return
		}
		if err != nil {
			log.Error("Failed to decode Request body", sl.Err(err))
			responseError(w, r, "Failed to decode request")
			return
		}
		log.Info("Reqeust body decoded", slog.Any("request", request))

		if err := validator.New().Struct(request); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("Invalid request", sl.Err(err))
			responseError(w, r, validateErr.Error())
			return
		}

		alias := request.Alias
		if len(alias) <= 0 {
			alias = random.GenerateRandomString(6)
		}

		id, err := urlSaver.Save(request.Url, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("URL already exists", slog.String("url", request.Url))
			responseError(w, r, "URL already exists")
			return
		}
		if err != nil {
			log.Error("Failed to add URL", sl.Err(err))
			responseError(w, r, "Failed to add URL")
			return
		}

		log.Info("URL added", slog.Int64("id", id))
		responseOk(w, r, alias)
	}
}

func responseOk(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		Alias:    alias,
	})
}

func responseError(w http.ResponseWriter, r *http.Request, errMsg string) {
	render.JSON(w, r, response.Error(errMsg))
}

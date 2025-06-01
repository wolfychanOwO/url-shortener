package save

import (
	"log/slog"
	"main/internal/lib/api/response"
	"main/internal/lib/logger/sl"
	"main/internal/lib/random"
	"main/internal/storage"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
    URL string `json:"url" validate:"required,url"`
    Alias string `json:"alias,omitempty"`
}

type Response struct {
    response.Response
    Alias string `json:"alias,omitempty"`
}

const aliasLength = 6

//go:generate go run github.com/vektra/mockery/v3 --name=URLSaver
type URLSaver interface {
    SaveURL(urlToSave, alias string) (int64, error)
}


func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        const op = "handlers.url.save.New"

        log = log.With(
            slog.String("op", op),
            slog.String("request_id", middleware.GetReqID(r.Context())),
        )

        var req Request

        err := render.DecodeJSON(r.Body, &req)
        if err != nil {
            log.Error("failed to decode request", sl.Err(err))

            render.JSON(w, r, response.ERROR("failed to decode request"))

            return 
        }

        log.Info("request body decoded", slog.Any("request", req))

        if err := validator.New().Struct(req) ; err != nil {
            validateErr := err.(validator.ValidationErrors)
            
            log.Error("invalid request", sl.Err(err))
            
            render.JSON(w, r, response.ERROR("invalid request"))
            render.JSON(w, r, response.ValidatationError(validateErr))

            return
        }

        alias := req.Alias
        if alias == "" {
            alias = random.NewRandomString(aliasLength)
        }

        id, err := urlSaver.SaveURL(req.URL, alias)
        switch err {
        case storage.ErrURLExists:
            log.Info("url already exists", slog.String("url", req.URL))

            render.JSON(w, r, response.ERROR("url already exists"))

            return
        case nil:
            log.Info("url added", slog.Int64("id", id))
            responseOK(w, r, alias)
        default:
            log.Error("failed to add url", sl.Err(err))

            render.JSON(w, r, response.ERROR("failed to add url"))

            return
        }
    }
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
    render.JSON(w, r, Response{
        Response: response.OK(),
        Alias: alias,
    })
}

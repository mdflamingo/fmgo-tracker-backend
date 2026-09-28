package apiresponse

import (
	"net/http"
	"runtime/debug"

	"github.com/go-chi/render"
	"github.com/mdflamingo/fgo-tracker-backend/internal/logger"
	"go.uber.org/zap"
)

type ResponseError struct {
	ErrorMsg string            `json:"error"`
	Code     int               `json:"code"`
	Details  map[string]string `json:"details,omitempty"`
}

func (e *ResponseError) Render(r *http.Request) error {
	return nil
}

func ResponseWithError(w http.ResponseWriter, r *http.Request, code int, errMsg string) {
	render.Status(r, code)
	render.JSON(w, r, &ResponseError{
		ErrorMsg: errMsg,
		Code:     code,
	})
}

func ResponseWithValidationError(w http.ResponseWriter, r *http.Request, details map[string]string) {
	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, &ResponseError{
		ErrorMsg: "Validation failed",
		Code:     http.StatusBadRequest,
		Details:  details,
	})
}

// Recoverer turns a panic in a downstream handler into a regular JSON 500, so
// every error this API returns has the same shape. chi's middleware.Recoverer
// only calls WriteHeader, which leaves the client with an empty body and sends
// the stack trace to stderr instead of the application log.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &recovererWriter{ResponseWriter: w}

		defer func() {
			rvr := recover()
			if rvr == nil {
				return
			}

			// http.ErrAbortHandler is the documented way to abandon a response
			// on purpose, so it must keep propagating through net/http.
			if rvr == http.ErrAbortHandler {
				panic(rvr)
			}

			logger.Log.Error("panic recovered",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Any("panic", rvr),
				zap.String("panic_stack", string(debug.Stack())),
			)

			// A response that already started cannot be turned into a 500, and
			// an upgrade handshake must not get a JSON body.
			if rw.wroteHeader || r.Header.Get("Connection") == "Upgrade" {
				return
			}

			ResponseWithError(rw, r, http.StatusInternalServerError, "Internal Server Error")
		}()

		next.ServeHTTP(rw, r)
	})
}

type recovererWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *recovererWriter) WriteHeader(status int) {
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

// Unwrap keeps http.ResponseController able to reach the underlying writer.
func (w *recovererWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Package httpx contains the HTTP helpers shared by all handler packages.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
	"github.com/TimoSto/tax-return-organizer/backend/internal/logging"
)

const (
	// RequestIDHeader carries the id that ties a response to its log lines.
	RequestIDHeader = "X-Request-ID"
	// MaxBodyBytes limits the size of request bodies handlers read.
	MaxBodyBytes = 1 << 20
)

// CollectorIDParam is the path parameter that carries the collector id in tenant-scoped patterns.
const CollectorIDParam = "collectorId"

// Route is an endpoint: a ServeMux pattern ("POST /collectors") and its handler.
type Route struct {
	Pattern string
	Handler http.Handler
}

// ProblemContentType is the media type of every error answer (RFC 9457).
const ProblemContentType = "application/problem+json"

// Problem is the body of every error answer, a problem details object as defined by RFC 9457.
// Clients switch on Code, which is stable; Detail is an English text for developers, not for end users.
type Problem struct {
	// Type identifies the problem class. It is a URI reference derived from Code and is not dereferenced.
	Type string `json:"type"`
	// Title is the HTTP reason phrase of Status, so it is the same for every occurrence of a class.
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
	// Code is the stable snake_case identifier of the problem, e.g. "collector_not_found".
	Code string `json:"code"`
	// RequestID repeats the X-Request-ID header, so it survives when only the body is kept.
	RequestID string `json:"requestId,omitempty"`
}

// Codes of the problems that do not originate from a domain error.
const (
	CodeInternal            = "internal_error"
	CodeInvalidBody         = "invalid_request_body"
	CodeBodyTooLarge        = "request_body_too_large"
	CodeUnsupportedMedia    = "unsupported_media_type"
	problemTypePrefix       = "/problems/"
	internalErrorDetailText = "internal error"
)

// WriteJSON writes body as JSON with the given status.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write response", "error", err)
	}
}

// WriteProblem answers with a problem details object. The request is used for the request id.
func WriteProblem(w http.ResponseWriter, r *http.Request, status int, code, detail string) {
	p := Problem{
		Type:   problemTypePrefix + strings.ReplaceAll(code, "_", "-"),
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
		Code:   code,
	}
	p.RequestID, _ = logging.RequestIDFrom(r.Context())

	w.Header().Set("Content-Type", ProblemContentType)
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(p); err != nil {
		slog.ErrorContext(r.Context(), "write response", "error", err)
	}
}

// Tenant returns the collector the request acts for, which handlers pass to the use cases. Today it is the
// {collectorId} path value; once there is authentication, only this function changes: the tenant then comes from
// the authenticated principal. On failure it answers (normally 400) and returns false.
func Tenant(w http.ResponseWriter, r *http.Request) (domain.CollectorID, bool) {
	return ParsePathID(w, r, CollectorIDParam, domain.ParseCollectorID)
}

// ParsePathID parses the path value param with parse. On failure it answers with the (domain) error parse
// returned, normally a 400, and returns false.
func ParsePathID[T any](w http.ResponseWriter, r *http.Request, param string, parse func(string) (T, error)) (T, bool) {
	id, err := parse(r.PathValue(param))
	if err != nil {
		WriteError(w, r, err)

		var zero T

		return zero, false
	}

	return id, true
}

// DecodeJSON reads the request body as exactly one JSON object into dst, rejecting unknown fields and trailing
// data. On failure it answers 415, 413 or 400 and returns false.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		WriteProblem(w, r, http.StatusUnsupportedMediaType, CodeUnsupportedMedia, "content type must be application/json")

		return false
	}

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err == nil {
		// A second value, or garbage after the first one, means this is not the single object we asked for.
		if _, e := dec.Token(); !errors.Is(e, io.EOF) {
			err = errors.New("unexpected data after the JSON object")
		}
	}

	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			WriteProblem(w, r, http.StatusRequestEntityTooLarge, CodeBodyTooLarge, "request body too large")

			return false
		}

		WriteProblem(w, r, http.StatusBadRequest, CodeInvalidBody, "invalid request body")

		return false
	}

	return true
}

// WriteError maps err to an HTTP status. Domain errors are reported to the client with their own code and
// message only (never the wrapped chain); everything else is logged and answered with a generic 500.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var de *domain.Error
	if errors.As(err, &de) {
		if code, ok := codeFor(de); ok {
			WriteProblem(w, r, statusFor(de.Kind), code, de.Error())

			return
		}

		// A domain error without a code is a programming error (see problemCodes), not the client's mistake.
		slog.ErrorContext(r.Context(), "domain error without a problem code", "error", err)
		WriteProblem(w, r, http.StatusInternalServerError, CodeInternal, internalErrorDetailText)

		return
	}

	slog.ErrorContext(r.Context(), "request failed", "error", err)
	WriteProblem(w, r, http.StatusInternalServerError, CodeInternal, internalErrorDetailText)
}

func statusFor(k domain.ErrorKind) int {
	switch k {
	case domain.ErrorInvalid:
		return http.StatusBadRequest
	case domain.ErrorNotFound:
		return http.StatusNotFound
	case domain.ErrorConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

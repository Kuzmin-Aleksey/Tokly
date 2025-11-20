package middlewarex

import (
	"FairLAP/pkg/logx"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
)

func RequestLogging(
	sensitiveDataMasker logx.SensitiveDataMaskerInterface,
	logFieldMaxLen int,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if strings.Contains(r.RequestURI, "image") ||
				strings.Contains(r.RequestURI, "static") ||
				strings.Contains(r.RequestURI, "ico") {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			dumpBody := true

			if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
				dumpBody = false
			}

			dump, err := httputil.DumpRequest(r, dumpBody)
			if err != nil {
				logger(ctx).Error("Failed to dump http request", logx.Error(err))
			}

			if len(dump) > logFieldMaxLen {
				dump = dump[:logFieldMaxLen]
			}

			logger(ctx).Info(
				logx.FieldHTTPRequest,
				slog.String(logx.FieldRequestBody, string(sensitiveDataMasker.Mask(dump))),
			)

			next.ServeHTTP(w, r)
		})
	}
}

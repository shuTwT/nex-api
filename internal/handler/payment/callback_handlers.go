package payment

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
)

func (h *Handler) wechatCallback(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	// The direct WeChat provider verifies the raw JSON body, while an epay
	// gateway notifies via query string or urlencoded form, so both are passed.
	_, err = h.service.HandleRawCallback(r.Context(), "wechat", body, r.Header, callbackForm(r, body))
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	writeRawJSON(w, http.StatusOK, `{"code":"SUCCESS","message":"成功"}`)
}

func (h *Handler) alipayCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.callbackError(w, r, err)
		return
	}
	// r.Form merges query and body parameters so epay GET notifications verify
	// the same way as POST ones.
	_, err := h.service.HandleRawCallback(r.Context(), "alipay", nil, nil, r.Form)
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	writeRaw(w, http.StatusOK, "text/plain; charset=utf-8", "success")
}

// callbackForm collects the request query and, when the body is form encoded,
// its fields. The body is consumed by the caller before parsing, so it is
// re-parsed from the buffered copy.
func callbackForm(r *http.Request, body []byte) map[string][]string {
	form := url.Values{}
	for key, values := range r.URL.Query() {
		form[key] = values
	}
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") || strings.HasPrefix(contentType, "multipart/form-data") {
		if values, err := url.ParseQuery(string(body)); err == nil {
			for key, value := range values {
				form[key] = append(form[key], value...)
			}
		}
	}
	return form
}

func (h *Handler) mockCallback(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	_, err = h.service.HandleRawCallback(r.Context(), "mock", body, nil, nil)
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, struct {
		Success bool `json:"success"`
	}{Success: true})
}

func (h *Handler) callbackError(w http.ResponseWriter, r *http.Request, err error) {
	handlerutils.WriteError(w, r, err)
}

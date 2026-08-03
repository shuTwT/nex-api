package catalog

import (
	"net/http"

	handlerutils "github.com/shuTwT/nex-api/backend/internal/pkg/utils"
	servicecatalog "github.com/shuTwT/nex-api/backend/internal/service/catalog"
)

func (h *Handler) listAPIs(w http.ResponseWriter, r *http.Request) {
	options, err := parseAPIListOptions(r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	result, err := h.apis.List(r.Context(), options)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	data := make([]apiData, len(result.Items))
	for index, item := range result.Items {
		data[index] = toAPIData(item, false)
	}
	handlerutils.WritePaginated(w, data, options.Page, options.Limit, result.Total)
}

func (h *Handler) createAPI(w http.ResponseWriter, r *http.Request) {
	var request apiRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	input := servicecatalog.APIInput{Name: request.Name, Alias: request.Alias, Description: request.Description, Endpoint: request.Endpoint, Method: request.Method, CategoryID: request.CategoryID, Documentation: request.Documentation, PreScript: request.PreScript, PostScript: request.PostScript, Pricing: intValue(request.Pricing), IsActive: boolValue(request.IsActive, true)}
	item, err := h.apis.Create(r.Context(), input)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, toAPIData(item, true))
}

func (h *Handler) getAPI(w http.ResponseWriter, r *http.Request) {
	item, err := h.apis.GetByIdentifier(r.Context(), r.PathValue("id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toAPIData(item, true))
}

func (h *Handler) updateAPI(w http.ResponseWriter, r *http.Request) {
	var request apiUpdateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	item, err := h.apis.Update(r.Context(), r.PathValue("id"), servicecatalog.APIUpdateInput{Name: request.Name, Alias: request.Alias, Description: request.Description, Endpoint: request.Endpoint, Method: request.Method, CategoryID: request.CategoryID, Pricing: request.Pricing, Documentation: request.Documentation, PreScript: request.PreScript, PostScript: request.PostScript, IsActive: request.IsActive})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toAPIData(item, false))
}

func (h *Handler) deleteAPI(w http.ResponseWriter, r *http.Request) {
	if err := h.apis.Delete(r.Context(), r.PathValue("id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "API 已删除"})
}

func (h *Handler) toggleAPI(w http.ResponseWriter, r *http.Request) {
	item, err := h.apis.Toggle(r.Context(), r.PathValue("id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toAPIData(item, false))
}

func (h *Handler) apiStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.apis.Stats(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
}

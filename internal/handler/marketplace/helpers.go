package marketplace

import (
	"fmt"
	"net/http"
	"strings"

	appRuntime "github.com/shuTwT/nex-api/internal/handler/httpkit"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type pageOptions struct {
	page     int
	limit    int
	search   string
	category string
}

type apiView = model.MarketplaceAPIResp
type mcpView = model.MarketplaceMCPResp

func parsePage(r *http.Request, defaultLimit, maxLimit int) (pageOptions, error) {
	query := r.URL.Query()
	page, err := handlerutils.PositiveInt(query.Get("page"), 1)
	if err != nil {
		return pageOptions{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "page", Reason: "must be a positive integer"})
	}
	limit, err := handlerutils.PositiveInt(query.Get("limit"), defaultLimit)
	if err != nil || limit > maxLimit {
		return pageOptions{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "limit", Reason: fmt.Sprintf("must be between 1 and %d", maxLimit)})
	}
	return pageOptions{page: page, limit: limit, search: strings.TrimSpace(query.Get("search")), category: strings.TrimSpace(query.Get("category"))}, nil
}

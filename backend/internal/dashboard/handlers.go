package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/api"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/apiusage"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/user"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
	"github.com/shuTwT/nex-api/backend/internal/stats"
)

type Handler struct {
	db    *ent.Client
	stats *stats.Store
}

func NewHandler(db *ent.Client, statStore *stats.Store) (*Handler, error) {
	if db == nil || statStore == nil {
		return nil, errors.New("dashboard: database and stats store are required")
	}
	return &Handler{db: db, stats: statStore}, nil
}

func RegisterRoutes(mux *http.ServeMux, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("dashboard: mux and handler are required")
	}
	mux.HandleFunc("GET /api/dashboard/stats", handler.dashboardStats)
	mux.HandleFunc("GET /api/dashboard/activity", handler.activity)
	mux.HandleFunc("GET /api/dashboard/top-apis", handler.topAPIs)
	mux.HandleFunc("GET /api/dashboard/usage-trend", handler.usageTrend)
	mux.HandleFunc("GET /api/usage", handler.usage)
	mux.HandleFunc("GET /api/stats", handler.globalStats)
	mux.HandleFunc("GET /api/stats/{alias}", handler.apiStats)
	return nil
}

type activityView struct {
	ID        string `json:"id"`
	APIName   string `json:"apiName"`
	APIAlias  string `json:"apiAlias"`
	Credits   int    `json:"credits"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type topAPIView struct {
	Name       string `json:"name"`
	Calls      int64  `json:"calls"`
	Percentage int    `json:"percentage"`
}

func (h *Handler) dashboardStats(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthly, err := monthlyCreditsUsed(r.Context(), h.db, principal.UserID, firstDay)
	if err != nil {
		writeError(w, r, fmt.Errorf("sum monthly credits: %w", err))
		return
	}
	snapshot, err := h.stats.Snapshot(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("read dashboard counters: %w", err))
		return
	}
	apiCalls := int64(0)
	active := 0
	for key, calls := range snapshot.UserAPIs {
		if key.UserID != principal.UserID {
			continue
		}
		apiCalls += calls
		if calls > 0 {
			active++
		}
	}
	account, err := h.db.User.Query().Where(user.ID(principal.UserID)).Select(user.FieldCredits).Only(r.Context())
	if err != nil && !ent.IsNotFound(err) {
		writeError(w, r, fmt.Errorf("load account balance: %w", err))
		return
	}
	balance := 0
	if account != nil {
		balance = account.Credits
	}
	writeData(w, http.StatusOK, map[string]int64{"monthlyCreditsUsed": int64(monthly), "apiCalls": apiCalls, "accountBalance": int64(balance), "activeApis": int64(active)})
}

func monthlyCreditsUsed(ctx context.Context, db *ent.Client, userID string, firstDay time.Time) (int, error) {
	return db.ApiUsage.Query().
		Where(apiusage.UserId(userID), apiusage.CreatedAtGTE(firstDay)).
		Aggregate(func(selector *sql.Selector) string {
			return "COALESCE(" + sql.Sum(selector.C(apiusage.FieldCredits)) + ", 0)"
		}).
		Int(ctx)
}

func (h *Handler) activity(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	items, err := h.db.ApiUsage.Query().Where(apiusage.UserId(principal.UserID)).WithAPI().Order(apiusage.ByCreatedAt(sql.OrderDesc())).Limit(10).All(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("list recent activity: %w", err))
		return
	}
	activities := make([]activityView, 0, len(items))
	for _, item := range items {
		name, alias := "", ""
		if item.Edges.API != nil {
			name, alias = item.Edges.API.Name, item.Edges.API.Alias
		}
		activities = append(activities, activityView{ID: item.ID, APIName: name, APIAlias: alias, Credits: item.Credits, Status: item.Status, CreatedAt: item.CreatedAt.UTC().Format(timeFormat)})
	}
	writeData(w, http.StatusOK, activities)
}

func (h *Handler) topAPIs(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	snapshot, err := h.stats.Snapshot(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("read top API counters: %w", err))
		return
	}
	type apiCount struct {
		alias string
		calls int64
	}
	counts := make([]apiCount, 0)
	for key, calls := range snapshot.UserAPIs {
		if key.UserID == principal.UserID {
			counts = append(counts, apiCount{alias: key.Alias, calls: calls})
		}
	}
	sort.Slice(counts, func(i, j int) bool { return counts[i].calls > counts[j].calls })
	if len(counts) > 5 {
		counts = counts[:5]
	}
	if len(counts) == 0 {
		writeData(w, http.StatusOK, []topAPIView{})
		return
	}
	total := int64(0)
	for _, item := range counts {
		total += item.calls
	}
	aliases := make([]string, 0, len(counts))
	for _, item := range counts {
		aliases = append(aliases, item.alias)
	}
	details, err := h.db.Api.Query().Where(api.AliasIn(aliases...)).All(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("load top API details: %w", err))
		return
	}
	names := make(map[string]string, len(details))
	for _, detail := range details {
		names[detail.Alias] = detail.Name
	}
	result := make([]topAPIView, 0, len(counts))
	for _, item := range counts {
		name := names[item.alias]
		if name == "" {
			name = item.alias
		}
		percentage := 0
		if total > 0 {
			percentage = int((item.calls*100 + total/2) / total)
		}
		result = append(result, topAPIView{Name: name, Calls: item.calls, Percentage: percentage})
	}
	writeData(w, http.StatusOK, result)
}

func (h *Handler) usageTrend(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	now := time.Now()
	labels := make([]string, 0, 7)
	for offset := 6; offset >= 0; offset-- {
		labels = append(labels, now.Add(-time.Duration(offset)*time.Hour).Format("15:04"))
	}
	userTrend, err := h.stats.HourlyUsageTrend(r.Context(), principal.UserID, 7)
	if err != nil {
		writeError(w, r, fmt.Errorf("read user usage trend: %w", err))
		return
	}
	datasets := []map[string]any{{"label": "我的用量", "data": userTrend, "borderColor": "rgb(59, 130, 246)", "backgroundColor": "rgba(59, 130, 246, 0.1)"}}
	if principal.Role == "admin" {
		globalTrend, trendErr := h.stats.HourlyUsageTrend(r.Context(), "", 7)
		if trendErr != nil {
			writeError(w, r, fmt.Errorf("read global usage trend: %w", trendErr))
			return
		}
		datasets = []map[string]any{{"label": "全局用量", "data": globalTrend, "borderColor": "rgb(59, 130, 246)", "backgroundColor": "rgba(59, 130, 246, 0.1)"}, {"label": "我的用量", "data": userTrend, "borderColor": "rgb(16, 185, 129)", "backgroundColor": "rgba(16, 185, 129, 0.1)"}}
	}
	writeData(w, http.StatusOK, map[string]any{"labels": labels, "datasets": datasets})
}

func (h *Handler) requireUser(w http.ResponseWriter, r *http.Request) (authz.Principal, bool) {
	principal, err := authz.UserPolicy(r.Context())
	if err != nil {
		status := http.StatusUnauthorized
		code := "unauthorized"
		message := "authentication required"
		if errors.Is(err, authz.ErrForbidden) {
			status = http.StatusForbidden
			code = "forbidden"
			message = "access denied"
		}
		writeError(w, r, runtime.NewAPIError(status, code, message, err))
		return authz.Principal{}, false
	}
	return principal, true
}

const timeFormat = "2006-01-02T15:04:05.000Z"

func writeData[T any](w http.ResponseWriter, status int, data T) {
	_ = runtime.WriteData(w, status, data)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) { _ = runtime.WriteError(w, r, err) }

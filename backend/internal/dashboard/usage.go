package dashboard

import (
	"fmt"
	"net/http"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/database/ent/apiusage"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/user"
)

func (h *Handler) usage(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	last7 := now.Add(-7 * 24 * time.Hour)
	last30 := now.Add(-30 * 24 * time.Hour)
	total, err := h.db.ApiUsage.Query().Count(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("count total usage: %w", err))
		return
	}
	count := func(start time.Time) (int, error) {
		return h.db.ApiUsage.Query().Where(apiusage.UserId(principal.UserID), apiusage.CreatedAtGTE(start)).Count(r.Context())
	}
	today, err := count(todayStart)
	if err != nil {
		writeError(w, r, fmt.Errorf("count today usage: %w", err))
		return
	}
	week, err := count(last7)
	if err != nil {
		writeError(w, r, fmt.Errorf("count week usage: %w", err))
		return
	}
	month, err := count(last30)
	if err != nil {
		writeError(w, r, fmt.Errorf("count month usage: %w", err))
		return
	}
	hourly := make([]int, 24)
	for hour := range 24 {
		start := todayStart.Add(time.Duration(hour) * time.Hour)
		end := start.Add(time.Hour)
		hourly[hour], err = h.db.ApiUsage.Query().Where(apiusage.UserId(principal.UserID), apiusage.CreatedAtGTE(start), apiusage.CreatedAtLT(end)).Count(r.Context())
		if err != nil {
			writeError(w, r, fmt.Errorf("count hourly usage: %w", err))
			return
		}
	}
	daily := func(start time.Time, days int) ([]int, error) {
		values := make([]int, days)
		for day := range values {
			from := start.Add(time.Duration(day) * 24 * time.Hour)
			values[day], err = h.db.ApiUsage.Query().Where(apiusage.UserId(principal.UserID), apiusage.CreatedAtGTE(from), apiusage.CreatedAtLT(from.Add(24*time.Hour))).Count(r.Context())
			if err != nil {
				return nil, err
			}
		}
		return values, nil
	}
	last7Daily, err := daily(last7, 7)
	if err != nil {
		writeError(w, r, fmt.Errorf("count daily week usage: %w", err))
		return
	}
	last30Daily, err := daily(last30, 30)
	if err != nil {
		writeError(w, r, fmt.Errorf("count daily month usage: %w", err))
		return
	}
	account, err := h.db.User.Query().Where(user.ID(principal.UserID)).Only(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("load usage balance: %w", err))
		return
	}
	writeData(w, http.StatusOK, map[string]any{"freeCredits": account.Credits, "totalUsage": total, "todayUsage": today, "last7DaysUsage": week, "last30DaysUsage": month, "todayHourlyUsage": hourly, "last7DaysDailyUsage": last7Daily, "last30DaysDailyUsage": last30Daily})
}

func (h *Handler) globalStats(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireUser(w, r); !ok {
		return
	}
	snapshot, err := h.stats.Snapshot(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("read global stats: %w", err))
		return
	}
	switch r.URL.Query().Get("type") {
	case "global":
		writeData(w, http.StatusOK, map[string]int64{"totalRequests": snapshot.GlobalRequests})
	case "all":
		writeData(w, http.StatusOK, snapshot.APIs)
	default:
		writeData(w, http.StatusOK, map[string]string{"message": "Specify type=global or type=all"})
	}
}

func (h *Handler) apiStats(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	alias := r.PathValue("alias")
	var count int64
	var err error
	if r.URL.Query().Get("user") == "true" {
		count, err = h.stats.UserAPIRequestCount(r.Context(), principal.UserID, alias)
	} else {
		count, err = h.stats.APIRequestCount(r.Context(), alias)
	}
	if err != nil {
		writeError(w, r, fmt.Errorf("read API stats: %w", err))
		return
	}
	writeData(w, http.StatusOK, map[string]any{"apiAlias": alias, "requestCount": count})
}

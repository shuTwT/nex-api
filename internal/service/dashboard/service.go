// Package dashboard provides the dashboard/usage statistics service.
package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"entgo.io/ent/dialect/sql"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/api"
	"github.com/shuTwT/nex-api/ent/apiusage"
	"github.com/shuTwT/nex-api/ent/user"
	"github.com/shuTwT/nex-api/internal/service/stats"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type ActivityView = model.DashboardActivityResp
type TopAPIView = model.DashboardTopAPIResp

// Service owns dashboard queries; the handler only adapts HTTP.
type Service struct {
	db    *ent.Client
	stats *stats.Store
}

func NewService(db *ent.Client, statStore *stats.Store) (*Service, error) {
	if db == nil || statStore == nil {
		return nil, fmt.Errorf("dashboard: database and stats store are required")
	}
	return &Service{db: db, stats: statStore}, nil
}

// DashboardStats returns the per-user dashboard counters.
func (s *Service) DashboardStats(ctx context.Context, userID string) (map[string]int64, error) {
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthly, err := monthlyCreditsUsed(ctx, s.db, userID, firstDay)
	if err != nil {
		return nil, fmt.Errorf("sum monthly credits: %w", err)
	}
	snapshot, err := s.stats.Snapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("read dashboard counters: %w", err)
	}
	apiCalls := int64(0)
	active := 0
	for key, calls := range snapshot.UserAPIs {
		if key.UserID != userID {
			continue
		}
		apiCalls += calls
		if calls > 0 {
			active++
		}
	}
	account, err := s.db.User.Query().Where(user.ID(userID)).Select(user.FieldCredits).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("load account balance: %w", err)
	}
	balance := 0
	if account != nil {
		balance = account.Credits
	}
	return map[string]int64{"monthlyCreditsUsed": int64(monthly), "apiCalls": apiCalls, "accountBalance": int64(balance), "activeApis": int64(active)}, nil
}

func monthlyCreditsUsed(ctx context.Context, db *ent.Client, userID string, firstDay time.Time) (int, error) {
	return db.ApiUsage.Query().
		Where(apiusage.UserId(userID), apiusage.CreatedAtGTE(firstDay)).
		Aggregate(func(selector *sql.Selector) string {
			return "COALESCE(" + sql.Sum(selector.C(apiusage.FieldCredits)) + ", 0)"
		}).
		Int(ctx)
}

// Activity returns the ten most recent usage rows for the user.
func (s *Service) Activity(ctx context.Context, userID string) ([]ActivityView, error) {
	items, err := s.db.ApiUsage.Query().Where(apiusage.UserId(userID)).WithAPI().Order(apiusage.ByCreatedAt(sql.OrderDesc())).Limit(10).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recent activity: %w", err)
	}
	activities := make([]ActivityView, 0, len(items))
	for _, item := range items {
		name, alias := "", ""
		if item.Edges.API != nil {
			name, alias = item.Edges.API.Name, item.Edges.API.Alias
		}
		activities = append(activities, ActivityView{ID: item.ID, APIName: name, APIAlias: alias, Credits: item.Credits, Status: item.Status, CreatedAt: item.CreatedAt.UTC().Format(timeFormat)})
	}
	return activities, nil
}

// TopAPIs returns the user's top five APIs by call count.
func (s *Service) TopAPIs(ctx context.Context, userID string) ([]TopAPIView, error) {
	snapshot, err := s.stats.Snapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("read top API counters: %w", err)
	}
	type apiCount struct {
		alias string
		calls int64
	}
	counts := make([]apiCount, 0)
	for key, calls := range snapshot.UserAPIs {
		if key.UserID == userID {
			counts = append(counts, apiCount{alias: key.Alias, calls: calls})
		}
	}
	sort.Slice(counts, func(i, j int) bool { return counts[i].calls > counts[j].calls })
	if len(counts) > 5 {
		counts = counts[:5]
	}
	if len(counts) == 0 {
		return []TopAPIView{}, nil
	}
	total := int64(0)
	for _, item := range counts {
		total += item.calls
	}
	aliases := make([]string, 0, len(counts))
	for _, item := range counts {
		aliases = append(aliases, item.alias)
	}
	details, err := s.db.Api.Query().Where(api.AliasIn(aliases...)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("load top API details: %w", err)
	}
	names := make(map[string]string, len(details))
	for _, detail := range details {
		names[detail.Alias] = detail.Name
	}
	result := make([]TopAPIView, 0, len(counts))
	for _, item := range counts {
		name := names[item.alias]
		if name == "" {
			name = item.alias
		}
		percentage := 0
		if total > 0 {
			percentage = int((item.calls*100 + total/2) / total)
		}
		result = append(result, TopAPIView{Name: name, Calls: item.calls, Percentage: percentage})
	}
	return result, nil
}

// UsageTrend returns hourly usage labels and datasets for the user.
func (s *Service) UsageTrend(ctx context.Context, userID, role string) (map[string]any, error) {
	now := time.Now()
	labels := make([]string, 0, 7)
	for offset := 6; offset >= 0; offset-- {
		labels = append(labels, now.Add(-time.Duration(offset)*time.Hour).Format("15:04"))
	}
	userTrend, err := s.stats.HourlyUsageTrend(ctx, userID, 7)
	if err != nil {
		return nil, fmt.Errorf("read user usage trend: %w", err)
	}
	datasets := []map[string]any{{"label": "我的用量", "data": userTrend, "borderColor": "rgb(59, 130, 246)", "backgroundColor": "rgba(59, 130, 246, 0.1)"}}
	if role == "admin" {
		globalTrend, trendErr := s.stats.HourlyUsageTrend(ctx, "", 7)
		if trendErr != nil {
			return nil, fmt.Errorf("read global usage trend: %w", trendErr)
		}
		datasets = []map[string]any{{"label": "全局用量", "data": globalTrend, "borderColor": "rgb(59, 130, 246)", "backgroundColor": "rgba(59, 130, 246, 0.1)"}, {"label": "我的用量", "data": userTrend, "borderColor": "rgb(16, 185, 129)", "backgroundColor": "rgba(16, 185, 129, 0.1)"}}
	}
	return map[string]any{"labels": labels, "datasets": datasets}, nil
}

// Usage returns the per-user usage summary.
func (s *Service) Usage(ctx context.Context, userID string) (map[string]any, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	last7 := now.Add(-7 * 24 * time.Hour)
	last30 := now.Add(-30 * 24 * time.Hour)
	total, err := s.db.ApiUsage.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count total usage: %w", err)
	}
	count := func(start time.Time) (int, error) {
		return s.db.ApiUsage.Query().Where(apiusage.UserId(userID), apiusage.CreatedAtGTE(start)).Count(ctx)
	}
	today, err := count(todayStart)
	if err != nil {
		return nil, fmt.Errorf("count today usage: %w", err)
	}
	week, err := count(last7)
	if err != nil {
		return nil, fmt.Errorf("count week usage: %w", err)
	}
	month, err := count(last30)
	if err != nil {
		return nil, fmt.Errorf("count month usage: %w", err)
	}
	hourly := make([]int, 24)
	for hour := range 24 {
		start := todayStart.Add(time.Duration(hour) * time.Hour)
		end := start.Add(time.Hour)
		hourly[hour], err = s.db.ApiUsage.Query().Where(apiusage.UserId(userID), apiusage.CreatedAtGTE(start), apiusage.CreatedAtLT(end)).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("count hourly usage: %w", err)
		}
	}
	daily := func(start time.Time, days int) ([]int, error) {
		values := make([]int, days)
		for day := range values {
			from := start.Add(time.Duration(day) * 24 * time.Hour)
			values[day], err = s.db.ApiUsage.Query().Where(apiusage.UserId(userID), apiusage.CreatedAtGTE(from), apiusage.CreatedAtLT(from.Add(24*time.Hour))).Count(ctx)
			if err != nil {
				return nil, err
			}
		}
		return values, nil
	}
	last7Daily, err := daily(last7, 7)
	if err != nil {
		return nil, fmt.Errorf("count daily week usage: %w", err)
	}
	last30Daily, err := daily(last30, 30)
	if err != nil {
		return nil, fmt.Errorf("count daily month usage: %w", err)
	}
	account, err := s.db.User.Query().Where(user.ID(userID)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("load usage balance: %w", err)
	}
	return map[string]any{"freeCredits": account.Credits, "totalUsage": total, "todayUsage": today, "last7DaysUsage": week, "last30DaysUsage": month, "todayHourlyUsage": hourly, "last7DaysDailyUsage": last7Daily, "last30DaysDailyUsage": last30Daily}, nil
}

// GlobalStats returns global or per-API request counters.
func (s *Service) GlobalStats(ctx context.Context, statsType string) (any, error) {
	snapshot, err := s.stats.Snapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("read global stats: %w", err)
	}
	switch statsType {
	case "global":
		return map[string]int64{"totalRequests": snapshot.GlobalRequests}, nil
	case "all":
		return snapshot.APIs, nil
	default:
		return map[string]string{"message": "Specify type=global or type=all"}, nil
	}
}

// APIStats returns the request count for one API alias.
func (s *Service) APIStats(ctx context.Context, userID, alias string, perUser bool) (map[string]any, error) {
	var count int64
	var err error
	if perUser {
		count, err = s.stats.UserAPIRequestCount(ctx, userID, alias)
	} else {
		count, err = s.stats.APIRequestCount(ctx, alias)
	}
	if err != nil {
		return nil, fmt.Errorf("read API stats: %w", err)
	}
	return map[string]any{"apiAlias": alias, "requestCount": count}, nil
}

const timeFormat = "2006-01-02T15:04:05.000Z"

package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/shuTwT/nex-api/ent"
)

// jsonFieldNames returns the JSON keys a struct emits, ignoring ent internals:
// the unexported config (json:"-"), the edges struct and the selectValues scan
// payload (ent relation/selection internals, not part of the public API).
func jsonFieldNames(t reflect.Type) map[string]struct{} {
	fields := make(map[string]struct{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name == "" {
			name = field.Name
		}
		if name == "edges" || name == "selectValues" {
			continue
		}
		fields[name] = struct{}{}
	}
	return fields
}

// TestEntityRespDTOsMatchEntities locks the JSON field sets of the
// hand-written Resp DTOs to the ent entities whose handler responses they
// document. Swaggo cannot resolve ent package types from handler annotations,
// so the spec references these DTOs; if an entity field is added or its JSON
// tag changes, this test fails until the matching DTO is updated.
func TestEntityRespDTOsMatchEntities(t *testing.T) {
	cases := []struct {
		entity any
		dto    any
	}{
		{ent.Advertisement{}, AdvertisementResp{}},
		{ent.SubscriptionPlan{}, SubscriptionPlanResp{}},
		{ent.Subscription{}, SubscriptionResp{}},
		{ent.RedemptionCode{}, RedemptionCodeResp{}},
		{ent.SystemSetting{}, SystemSettingResp{}},
	}

	for _, tc := range cases {
		entityFields := jsonFieldNames(reflect.TypeOf(tc.entity))
		dtoFields := jsonFieldNames(reflect.TypeOf(tc.dto))
		var missing, extra []string
		for name := range entityFields {
			if _, ok := dtoFields[name]; !ok {
				missing = append(missing, name)
			}
		}
		for name := range dtoFields {
			if _, ok := entityFields[name]; !ok {
				extra = append(extra, name)
			}
		}
		if len(missing) > 0 || len(extra) > 0 {
			t.Errorf("%T vs %T diverged: missing=%v extra=%v", tc.entity, tc.dto, missing, extra)
		}
	}
}

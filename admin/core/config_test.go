package core

import (
	"testing"
)

func TestConfig_ListDisplay(t *testing.T) {
	tests := []struct {
		name        string
		listDisplay []Field
	}{
		{
			name:        "empty list display",
			listDisplay: []Field{},
		},
		{
			name:        "single field",
			listDisplay: []Field{Method("name")},
		},
		{
			name:        "multiple fields",
			listDisplay: []Field{Method("name"), Method("email"), Method("created_at")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config[struct{}]{
				ListDisplay: tt.listDisplay,
			}
			if len(config.ListDisplay) != len(tt.listDisplay) {
				t.Errorf("Config.ListDisplay length = %d, want %d", len(config.ListDisplay), len(tt.listDisplay))
			}
		})
	}
}

func TestConfig_SearchFields(t *testing.T) {
	searchFields := []Field{Method("name"), Method("email")}
	config := Config[struct{}]{
		SearchFields: searchFields,
	}
	if len(config.SearchFields) != 2 {
		t.Errorf("Config.SearchFields length = %d, want 2", len(config.SearchFields))
	}
}

func TestConfig_Ordering(t *testing.T) {
	ordering := []Field{Method("-created_at"), Method("name")}
	config := Config[struct{}]{
		Ordering: ordering,
	}
	if len(config.Ordering) != 2 {
		t.Errorf("Config.Ordering length = %d, want 2", len(config.Ordering))
	}
}

func TestConfig_ListPerPage(t *testing.T) {
	tests := []struct {
		name    string
		perPage int
	}{
		{"default", 0},
		{"custom 25", 25},
		{"custom 100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config[struct{}]{
				ListPerPage: tt.perPage,
			}
			if config.ListPerPage != tt.perPage {
				t.Errorf("Config.ListPerPage = %d, want %d", config.ListPerPage, tt.perPage)
			}
		})
	}
}

func TestConfig_PageType(t *testing.T) {
	tests := []struct {
		name     string
		pageType string
	}{
		{"default", ""},
		{"list", "list"},
		{"form", "form"},
		{"list-form", "list-form"},
		{"detail", "detail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config[struct{}]{
				PageType: tt.pageType,
			}
			if config.PageType != tt.pageType {
				t.Errorf("Config.PageType = %q, want %q", config.PageType, tt.pageType)
			}
		})
	}
}

func TestConfig_VerboseName(t *testing.T) {
	config := Config[struct{}]{
		VerboseName:       "User",
		VerboseNamePlural: "Users",
		Description:       "User account management",
	}
	if config.VerboseName != "User" {
		t.Errorf("Config.VerboseName = %q, want %q", config.VerboseName, "User")
	}
	if config.VerboseNamePlural != "Users" {
		t.Errorf("Config.VerboseNamePlural = %q, want %q", config.VerboseNamePlural, "Users")
	}
	if config.Description != "User account management" {
		t.Errorf("Config.Description = %q, want %q", config.Description, "User account management")
	}
}

func TestConfig_Fields(t *testing.T) {
	fields := []string{"username", "email", "first_name", "last_name"}
	config := Config[struct{}]{
		Fields: fields,
	}
	if len(config.Fields) != 4 {
		t.Errorf("Config.Fields length = %d, want 4", len(config.Fields))
	}
}

func TestConfig_Exclude(t *testing.T) {
	exclude := []string{"password", "salt"}
	config := Config[struct{}]{
		Exclude: exclude,
	}
	if len(config.Exclude) != 2 {
		t.Errorf("Config.Exclude length = %d, want 2", len(config.Exclude))
	}
}

func TestConfig_ReadOnlyFields(t *testing.T) {
	readOnly := []string{"created_at", "updated_at"}
	config := Config[struct{}]{
		ReadOnlyFields: readOnly,
	}
	if len(config.ReadOnlyFields) != 2 {
		t.Errorf("Config.ReadOnlyFields length = %d, want 2", len(config.ReadOnlyFields))
	}
}

func TestConfig_Fieldsets(t *testing.T) {
	fieldsets := []Fieldset[struct{}]{
		{Name: "Personal Info", Fields: []string{"first_name", "last_name", "email"}},
		{Name: "Permissions", Fields: []string{"is_active", "is_staff"}},
	}
	config := Config[struct{}]{
		Fieldsets: fieldsets,
	}
	if len(config.Fieldsets) != 2 {
		t.Errorf("Config.Fieldsets length = %d, want 2", len(config.Fieldsets))
	}
}

func TestConfig_Actions(t *testing.T) {
	actions := []Action[struct{}]{
		{Name: "activate", Label: "Activate"},
		{Name: "deactivate", Label: "Deactivate"},
	}
	config := Config[struct{}]{
		Actions: actions,
	}
	if len(config.Actions) != 2 {
		t.Errorf("Config.Actions length = %d, want 2", len(config.Actions))
	}
}

func TestConfig_Filters(t *testing.T) {
	filters := []Filter[struct{}]{
		{Name: "status", Type: "choice"},
		{Name: "created", Type: "date_range"},
	}
	config := Config[struct{}]{
		Filters: filters,
	}
	if len(config.Filters) != 2 {
		t.Errorf("Config.Filters length = %d, want 2", len(config.Filters))
	}
}

func TestConfig_SaveAs(t *testing.T) {
	config := Config[struct{}]{
		SaveAs:         true,
		SaveAsContinue: true,
		SaveOnTop:      true,
	}
	if !config.SaveAs {
		t.Error("Config.SaveAs should be true")
	}
	if !config.SaveAsContinue {
		t.Error("Config.SaveAsContinue should be true")
	}
	if !config.SaveOnTop {
		t.Error("Config.SaveOnTop should be true")
	}
}

func TestConfig_ShowFullResultCount(t *testing.T) {
	config := Config[struct{}]{
		ShowFullResultCount: true,
	}
	if !config.ShowFullResultCount {
		t.Error("Config.ShowFullResultCount should be true")
	}
}

func TestConfig_PreserveFilters(t *testing.T) {
	config := Config[struct{}]{
		PreserveFilters: true,
	}
	if !config.PreserveFilters {
		t.Error("Config.PreserveFilters should be true")
	}
}

func TestConfig_EmptyValueDisplay(t *testing.T) {
	config := Config[struct{}]{
		EmptyValueDisplay: "-",
	}
	if config.EmptyValueDisplay != "-" {
		t.Errorf("Config.EmptyValueDisplay = %q, want %q", config.EmptyValueDisplay, "-")
	}
}

func TestConfig_DateHierarchy(t *testing.T) {
	config := Config[struct{}]{
		DateHierarchy: "created_at",
	}
	if config.DateHierarchy != "created_at" {
		t.Errorf("Config.DateHierarchy = %q, want %q", config.DateHierarchy, "created_at")
	}
}

func TestConfig_Icon(t *testing.T) {
	config := Config[struct{}]{
		Icon: "user",
	}
	if config.Icon != "user" {
		t.Errorf("Config.Icon = %q, want %q", config.Icon, "user")
	}
}

func TestConfig_UIOverrides(t *testing.T) {
	overrides := map[string]string{
		"list.actions": "CustomActions",
		"form.footer":  "CustomFooter",
	}
	config := Config[struct{}]{
		UIOverrides: overrides,
	}
	if len(config.UIOverrides) != 2 {
		t.Errorf("Config.UIOverrides length = %d, want 2", len(config.UIOverrides))
	}
}

func TestConfig_ListSelectRelated(t *testing.T) {
	related := []string{"author", "category"}
	config := Config[struct{}]{
		ListSelectRelated: related,
	}
	if len(config.ListSelectRelated) != 2 {
		t.Errorf("Config.ListSelectRelated length = %d, want 2", len(config.ListSelectRelated))
	}
}

func TestConfig_ListPrefetchRelated(t *testing.T) {
	prefetch := []string{"tags", "comments"}
	config := Config[struct{}]{
		ListPrefetchRelated: prefetch,
	}
	if len(config.ListPrefetchRelated) != 2 {
		t.Errorf("Config.ListPrefetchRelated length = %d, want 2", len(config.ListPrefetchRelated))
	}
}

func TestConfig_ListMaxShowAll(t *testing.T) {
	config := Config[struct{}]{
		ListMaxShowAll: 200,
	}
	if config.ListMaxShowAll != 200 {
		t.Errorf("Config.ListMaxShowAll = %d, want 200", config.ListMaxShowAll)
	}
}

func TestConfig_SearchHelpText(t *testing.T) {
	config := Config[struct{}]{
		SearchHelpText: "Search by name or email",
	}
	if config.SearchHelpText != "Search by name or email" {
		t.Errorf("Config.SearchHelpText = %q, want %q", config.SearchHelpText, "Search by name or email")
	}
}

func TestConfig_SortableBy(t *testing.T) {
	sortable := []Field{Method("name"), Method("created_at")}
	config := Config[struct{}]{
		SortableBy: sortable,
	}
	if len(config.SortableBy) != 2 {
		t.Errorf("Config.SortableBy length = %d, want 2", len(config.SortableBy))
	}
}

func TestConfig_ListDisplayLinks(t *testing.T) {
	links := []Field{Method("name")}
	config := Config[struct{}]{
		ListDisplayLinks: links,
	}
	if len(config.ListDisplayLinks) != 1 {
		t.Errorf("Config.ListDisplayLinks length = %d, want 1", len(config.ListDisplayLinks))
	}
}

func TestConfig_ListEditable(t *testing.T) {
	editable := []Field{Method("status"), Method("priority")}
	config := Config[struct{}]{
		ListEditable: editable,
	}
	if len(config.ListEditable) != 2 {
		t.Errorf("Config.ListEditable length = %d, want 2", len(config.ListEditable))
	}
}

func TestConfig_ListFilter(t *testing.T) {
	filters := []Field{Method("status"), Method("category")}
	config := Config[struct{}]{
		ListFilter: filters,
	}
	if len(config.ListFilter) != 2 {
		t.Errorf("Config.ListFilter length = %d, want 2", len(config.ListFilter))
	}
}

func TestConfig_RawIDFields(t *testing.T) {
	fields := []string{"author", "category"}
	config := Config[struct{}]{
		RawIDFields: fields,
	}
	if len(config.RawIDFields) != 2 {
		t.Errorf("Config.RawIDFields length = %d, want 2", len(config.RawIDFields))
	}
}

func TestConfig_AutocompleteFields(t *testing.T) {
	fields := []string{"author", "tags"}
	config := Config[struct{}]{
		AutocompleteFields: fields,
	}
	if len(config.AutocompleteFields) != 2 {
		t.Errorf("Config.AutocompleteFields length = %d, want 2", len(config.AutocompleteFields))
	}
}

func TestConfig_RadioFields(t *testing.T) {
	radioFields := map[string]RadioLayout{
		"status":  RadioHorizontal,
		"visible": RadioVertical,
	}
	config := Config[struct{}]{
		RadioFields: radioFields,
	}
	if len(config.RadioFields) != 2 {
		t.Errorf("Config.RadioFields length = %d, want 2", len(config.RadioFields))
	}
}

func TestConfig_PrepopulatedFields(t *testing.T) {
	prepopulated := map[string][]string{
		"slug": {"title"},
	}
	config := Config[struct{}]{
		PrepopulatedFields: prepopulated,
	}
	if len(config.PrepopulatedFields) != 1 {
		t.Errorf("Config.PrepopulatedFields length = %d, want 1", len(config.PrepopulatedFields))
	}
}

func TestConfig_InlineRelations(t *testing.T) {
	relations := map[string]InlineRelationConfig{
		"items": {Type: "one_to_many", Label: "Items"},
	}
	config := Config[struct{}]{
		InlineRelations: relations,
	}
	if len(config.InlineRelations) != 1 {
		t.Errorf("Config.InlineRelations length = %d, want 1", len(config.InlineRelations))
	}
}

func TestConfig_ShowChangeLink(t *testing.T) {
	config := Config[struct{}]{
		ShowChangeLink: true,
	}
	if !config.ShowChangeLink {
		t.Error("Config.ShowChangeLink should be true")
	}
}

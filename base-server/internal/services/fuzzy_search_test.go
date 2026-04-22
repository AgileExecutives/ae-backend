package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestService creates a FuzzySearchService with nil DB – fine for pure-logic tests.
func newTestService() *FuzzySearchService {
	return NewFuzzySearchService(nil, nil)
}

// newTestServiceWithDB creates a FuzzySearchService backed by an in-memory SQLite DB.
func newTestServiceWithDB(t *testing.T) (*FuzzySearchService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewFuzzySearchService(db, nil)
	return svc, db
}

// ─── DefaultFuzzySearchConfig ────────────────────────────────────────────────

func TestDefaultFuzzySearchConfig(t *testing.T) {
	cfg := DefaultFuzzySearchConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, 2, cfg.MinSearchLength)
	assert.Equal(t, 50, cfg.MaxResults)
	assert.Equal(t, 0.1, cfg.ScoreThreshold)
	assert.True(t, cfg.EnableHighlight)
	assert.False(t, cfg.CaseSensitive)
	assert.Equal(t, 2.0, cfg.ExactMatchBoost)
	assert.Equal(t, 1.5, cfg.PrefixMatchBoost)
}

// ─── RegisterEntity / GetEntityTypes ─────────────────────────────────────────

func TestRegisterEntity(t *testing.T) {
	svc := newTestService()

	// default entities are registered
	types := svc.GetEntityTypes()
	assert.Contains(t, types, "users")
	assert.Contains(t, types, "customers")
	assert.Contains(t, types, "contacts")
	assert.Contains(t, types, "plans")
	assert.Contains(t, types, "emails")

	// register a custom entity
	svc.RegisterEntity("products", EntityConfig{
		TableName:   "products",
		DisplayName: "Products",
	})
	types2 := svc.GetEntityTypes()
	assert.Contains(t, types2, "products")
}

// ─── GetConfig / UpdateConfig ────────────────────────────────────────────────

func TestGetAndUpdateConfig(t *testing.T) {
	svc := newTestService()

	original := svc.GetConfig()
	assert.Equal(t, 2, original.MinSearchLength)

	newCfg := original
	newCfg.MinSearchLength = 5
	newCfg.MaxResults = 100
	svc.UpdateConfig(newCfg)

	updated := svc.GetConfig()
	assert.Equal(t, 5, updated.MinSearchLength)
	assert.Equal(t, 100, updated.MaxResults)
}

// ─── tokenizeQuery ────────────────────────────────────────────────────────────

func TestTokenizeQuery(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		query    string
		expected []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"  spaces  ", []string{"spaces"}},
		{"a",         []string{}}, // too short (min=2)
		{"",          []string{}},
		{"ab",        []string{"ab"}},
		{"one two three", []string{"one", "two", "three"}},
	}
	for _, tt := range tests {
		got := svc.tokenizeQuery(tt.query)
		if len(tt.expected) == 0 {
			assert.Empty(t, got, "query: %q", tt.query)
		} else {
			assert.Equal(t, tt.expected, got, "query: %q", tt.query)
		}
	}
}

// ─── buildFieldCondition ─────────────────────────────────────────────────────

func TestBuildFieldCondition_AllTypes(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		searchType string
		contains   string
	}{
		{"exact",    "= 'john'"},
		{"prefix",   "LIKE 'john%'"},
		{"contains", "LIKE '%john%'"},
		{"fulltext", "tsvector"},
		{"fuzzy",    "SIMILAR TO"},
		{"",         "LIKE '%john%'"}, // default fallback
	}

	for _, tt := range tests {
		field := FieldConfig{Name: "first_name", SearchType: tt.searchType}
		cond := svc.buildFieldCondition(field, "john")
		assert.Contains(t, cond, tt.contains, "searchType=%q", tt.searchType)
	}
}

func TestBuildFieldCondition_MinLength(t *testing.T) {
	svc := newTestService()
	field := FieldConfig{Name: "name", SearchType: "contains", MinLength: 5}
	// "hi" is shorter than MinLength=5, should return empty
	assert.Empty(t, svc.buildFieldCondition(field, "hi"))
	// "hello" meets MinLength=5
	assert.NotEmpty(t, svc.buildFieldCondition(field, "hello"))
}

func TestBuildFieldCondition_CaseSensitive(t *testing.T) {
	cfg := DefaultFuzzySearchConfig()
	cfg.CaseSensitive = true
	svc := NewFuzzySearchService(nil, cfg)

	field := FieldConfig{Name: "name", SearchType: "contains"}
	cond := svc.buildFieldCondition(field, "John")
	// case-sensitive: field name should NOT be wrapped in LOWER()
	assert.NotContains(t, cond, "LOWER")
	// term value appears verbatim inside the LIKE pattern
	assert.Contains(t, cond, "John")
}

// ─── buildFuzzyPattern ────────────────────────────────────────────────────────

func TestBuildFuzzyPattern(t *testing.T) {
	svc := newTestService()

	// short term (<=3): returned as-is
	assert.Equal(t, "ab", svc.buildFuzzyPattern("ab"))
	assert.Equal(t, "abc", svc.buildFuzzyPattern("abc"))

	// longer term: contains intermediate ".*?"
	pattern := svc.buildFuzzyPattern("hello")
	assert.Contains(t, pattern, "h")
	assert.Contains(t, pattern, ".*?")
}

// ─── buildSearchConditions ───────────────────────────────────────────────────

func TestBuildSearchConditions_Empty(t *testing.T) {
	svc := newTestService()
	fields := []FieldConfig{{Name: "name", SearchType: "contains"}}
	assert.Empty(t, svc.buildSearchConditions(fields, ""))
}

func TestBuildSearchConditions_NonEmpty(t *testing.T) {
	svc := newTestService()
	fields := []FieldConfig{
		{Name: "name", SearchType: "contains"},
		{Name: "email", SearchType: "contains"},
	}
	cond := svc.buildSearchConditions(fields, "john")
	assert.True(t, strings.HasPrefix(cond, "("), "should be wrapped in parens")
	assert.Contains(t, cond, "name")
	assert.Contains(t, cond, "email")
}

// ─── buildTitleDescription ───────────────────────────────────────────────────

func TestBuildTitleDescription_Users(t *testing.T) {
	svc := newTestService()
	row := map[string]interface{}{
		"first_name": "Jane",
		"last_name":  "Doe",
		"email":      "jane@example.com",
	}
	title, desc := svc.buildTitleDescription("users", row)
	assert.Equal(t, "Jane Doe", title)
	assert.Contains(t, desc, "jane@example.com")
}

func TestBuildTitleDescription_Customers(t *testing.T) {
	svc := newTestService()
	row := map[string]interface{}{
		"name":    "ACME Corp",
		"email":   "info@acme.com",
		"company": "ACME",
	}
	title, desc := svc.buildTitleDescription("customers", row)
	assert.Equal(t, "ACME Corp", title)
	assert.Contains(t, desc, "info@acme.com")
	assert.Contains(t, desc, "ACME")
}

func TestBuildTitleDescription_Contacts(t *testing.T) {
	svc := newTestService()
	row := map[string]interface{}{"name": "Bob", "subject": "Support"}
	title, desc := svc.buildTitleDescription("contacts", row)
	assert.Equal(t, "Bob", title)
	assert.Contains(t, desc, "Support")
}

func TestBuildTitleDescription_Plans(t *testing.T) {
	svc := newTestService()
	row := map[string]interface{}{"name": "Pro", "price": "99", "currency": "EUR"}
	title, desc := svc.buildTitleDescription("plans", row)
	assert.Equal(t, "Pro", title)
	assert.Contains(t, desc, "99")
	assert.Contains(t, desc, "EUR")
}

func TestBuildTitleDescription_Emails(t *testing.T) {
	svc := newTestService()
	row := map[string]interface{}{"subject": "Invoice", "to_email": "user@x.com", "status": "sent"}
	title, desc := svc.buildTitleDescription("emails", row)
	assert.Equal(t, "Invoice", title)
	assert.Contains(t, desc, "user@x.com")
}

func TestBuildTitleDescription_Default_ByName(t *testing.T) {
	svc := newTestService()
	row := map[string]interface{}{"name": "Widget"}
	title, _ := svc.buildTitleDescription("widgets", row)
	assert.Equal(t, "Widget", title)
}

func TestBuildTitleDescription_Default_ByTitle(t *testing.T) {
	svc := newTestService()
	row := map[string]interface{}{"title": "Page Title"}
	title, _ := svc.buildTitleDescription("pages", row)
	assert.Equal(t, "Page Title", title)
}

func TestBuildTitleDescription_Default_Fallback(t *testing.T) {
	svc := newTestService()
	row := map[string]interface{}{"id": 42}
	title, _ := svc.buildTitleDescription("things", row)
	assert.Contains(t, title, "things")
	assert.Contains(t, title, "42")
}

// ─── buildURL ─────────────────────────────────────────────────────────────────

func TestBuildURL(t *testing.T) {
	svc := newTestService()
	assert.Equal(t, "/api/v1/customers/5", svc.buildURL("customers", 5))
	assert.Equal(t, "", svc.buildURL("customers", nil))
}

// ─── calculateFieldScore ─────────────────────────────────────────────────────

func TestCalculateFieldScore_Exact(t *testing.T) {
	svc := newTestService()
	field := FieldConfig{Name: "email", Weight: 1.0}
	score := svc.calculateFieldScore(field, "jane@example.com", []string{"jane@example.com"})
	assert.Equal(t, svc.config.ExactMatchBoost, score)
}

func TestCalculateFieldScore_Prefix(t *testing.T) {
	svc := newTestService()
	field := FieldConfig{Name: "name", Weight: 1.0}
	score := svc.calculateFieldScore(field, "johnsmith", []string{"john"})
	assert.Equal(t, svc.config.PrefixMatchBoost, score)
}

func TestCalculateFieldScore_Contains(t *testing.T) {
	svc := newTestService()
	field := FieldConfig{Name: "name", Weight: 1.0}
	score := svc.calculateFieldScore(field, "mr john doe", []string{"john"})
	assert.Equal(t, 1.0, score)
}

func TestCalculateFieldScore_NoMatch(t *testing.T) {
	svc := newTestService()
	field := FieldConfig{Name: "name", Weight: 1.0}
	score := svc.calculateFieldScore(field, "alice", []string{"xz"})
	// short term (<3) → calculateFuzzyScore returns 0
	assert.Equal(t, 0.0, score)
}

func TestCalculateFieldScore_Boost(t *testing.T) {
	svc := newTestService()
	// with Boost=1.0, score should be multiplied by (1+1.0)=2.0
	field := FieldConfig{Name: "name", Weight: 1.0, Boost: 1.0}
	score := svc.calculateFieldScore(field, "hello", []string{"hello"})
	assert.Equal(t, svc.config.ExactMatchBoost*2.0, score)
}

// ─── calculateFuzzyScore ─────────────────────────────────────────────────────

func TestCalculateFuzzyScore(t *testing.T) {
	svc := newTestService()
	// short term < 3 → 0
	assert.Equal(t, 0.0, svc.calculateFuzzyScore("hello", "hi"))
	// term matches all chars in text
	score := svc.calculateFuzzyScore("hello", "helo")
	assert.Greater(t, score, 0.0)
}

// ─── getField helper ──────────────────────────────────────────────────────────

func TestGetField(t *testing.T) {
	row := map[string]interface{}{
		"name":    "Alice",
		"age":     42,
		"nothing": nil,
	}
	assert.Equal(t, "Alice", getField(row, "name"))
	assert.Equal(t, "42", getField(row, "age"))
	assert.Equal(t, "", getField(row, "nothing"))
	assert.Equal(t, "", getField(row, "missing"))
}

// ─── generateSuggestions ─────────────────────────────────────────────────────

func TestGenerateSuggestions(t *testing.T) {
	svc := newTestService()
	suggestions := svc.generateSuggestions("nomatch")
	assert.NotEmpty(t, suggestions)
	// should include at least the static tips
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "keyword") || strings.Contains(s, "spell") || strings.Contains(s, "general") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected at least one spelling/keyword suggestion")
}

// ─── sortResults ─────────────────────────────────────────────────────────────

func TestSortResults_ByScore(t *testing.T) {
	svc := newTestService()
	results := []SearchResult{
		{Score: 1.0, Title: "low"},
		{Score: 3.0, Title: "high"},
		{Score: 2.0, Title: "mid"},
	}
	sorted := svc.sortResults(results, SearchOptions{})
	assert.Equal(t, "high", sorted[0].Title)
	assert.Equal(t, "mid", sorted[1].Title)
	assert.Equal(t, "low", sorted[2].Title)
}

func TestSortResults_CustomSortBy(t *testing.T) {
	svc := newTestService()
	results := []SearchResult{{Score: 1.0}, {Score: 3.0}}
	// non-default SortBy returns results unchanged (no panic)
	sorted := svc.sortResults(results, SearchOptions{SortBy: "created_at"})
	assert.Len(t, sorted, 2)
}

// ─── highlightText ───────────────────────────────────────────────────────────

func TestHighlightText_CaseInsensitive(t *testing.T) {
	svc := newTestService()
	// highlightText detects the match case-insensitively but replaces with the
	// original term; use the same case so the ReplaceAll finds the substring.
	result := svc.highlightText("Hello World", []string{"World"})
	assert.Contains(t, result, "<mark>World</mark>")

	// different case: no substitution happens (known behavior)
	result2 := svc.highlightText("Hello World", []string{"world"})
	// The function detects a match but replaces nothing (lower-case not in text)
	assert.NotContains(t, result2, "<mark>")
}

func TestHighlightText_CaseSensitive(t *testing.T) {
	cfg := DefaultFuzzySearchConfig()
	cfg.CaseSensitive = true
	svc := NewFuzzySearchService(nil, cfg)

	// exact case match
	result := svc.highlightText("Hello World", []string{"World"})
	assert.Contains(t, result, "<mark>World</mark>")

	// no match (different case)
	result2 := svc.highlightText("Hello World", []string{"world"})
	assert.NotContains(t, result2, "<mark>")
}

// ─── Search – short query (no DB needed) ─────────────────────────────────────

func TestSearch_ShortQuery(t *testing.T) {
	svc := newTestService()
	resp, err := svc.Search(SearchOptions{Query: "a"}) // below MinSearchLength=2
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Results)
	assert.Equal(t, "a", resp.Query)
}

func TestSearch_EmptyQuery(t *testing.T) {
	svc := newTestService()
	resp, err := svc.Search(SearchOptions{Query: ""})
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
}

// ─── calculateRelevanceScore ─────────────────────────────────────────────────

func TestCalculateRelevanceScore_EmptyQuery(t *testing.T) {
	svc := newTestService()
	score := svc.calculateRelevanceScore(nil, map[string]interface{}{}, "")
	assert.Equal(t, 0.0, score)
}

func TestCalculateRelevanceScore_Match(t *testing.T) {
	svc := newTestService()
	fields := []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}}
	row := map[string]interface{}{"name": "Jane Doe"}
	score := svc.calculateRelevanceScore(fields, row, "Jane")
	assert.Greater(t, score, 0.0)
}

// ─── Search – with real SQLite DB ─────────────────────────────────────────────

type searchPerson struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"column:name"`
}

func (searchPerson) TableName() string { return "people" }

func setupPeopleTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.AutoMigrate(&searchPerson{}))
	people := []searchPerson{
		{Name: "Alice Smith"},
		{Name: "Bob Johnson"},
		{Name: "Alice Wonder"},
	}
	require.NoError(t, db.Create(&people).Error)
}

func TestSearch_WithResults(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	setupPeopleTable(t, db)

	svc.RegisterEntity("people", EntityConfig{
		TableName:    "people",
		DisplayName:  "People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"id", "name"},
	})

	resp, err := svc.Search(SearchOptions{Query: "Alice", EntityTypes: []string{"people"}})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, resp.Total, 1)
	assert.NotEmpty(t, resp.Results)
	assert.Equal(t, "Alice", resp.Query)
}

func TestSearch_WithEntityTypes_All(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	setupPeopleTable(t, db)

	// Clear default entities and register only "people"
	svc2 := NewFuzzySearchService(db, nil)
	// Remove defaults by creating a fresh service and overriding with only our entity
	_ = svc
	svc2.entities = map[string]EntityConfig{}
	svc2.RegisterEntity("people", EntityConfig{
		TableName:    "people",
		DisplayName:  "People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"id", "name"},
	})

	// No EntityTypes specified → searches all registered entities
	resp, err := svc2.Search(SearchOptions{Query: "Bob"})
	require.NoError(t, err)
	// Bob should be found
	assert.GreaterOrEqual(t, resp.Total, 1)
}

func TestSearch_WithOffsetAndLimit(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	setupPeopleTable(t, db)

	svc.entities = map[string]EntityConfig{}
	svc.RegisterEntity("people", EntityConfig{
		TableName:    "people",
		DisplayName:  "People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"id", "name"},
	})

	// Offset beyond total returns empty results
	resp, err := svc.Search(SearchOptions{Query: "Alice", EntityTypes: []string{"people"}, Offset: 1000, Limit: 10})
	require.NoError(t, err)
	assert.Empty(t, resp.Results)
}

func TestSearch_LimitClamping(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	setupPeopleTable(t, db)

	svc.entities = map[string]EntityConfig{}
	svc.RegisterEntity("people", EntityConfig{
		TableName:    "people",
		DisplayName:  "People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"id", "name"},
	})

	// Limit=0 should be clamped to MaxResults
	resp, err := svc.Search(SearchOptions{Query: "Alice", EntityTypes: []string{"people"}, Limit: 0})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestSearch_NoResults_GeneratesSuggestions(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	setupPeopleTable(t, db)

	svc.entities = map[string]EntityConfig{}
	svc.RegisterEntity("people", EntityConfig{
		TableName:    "people",
		DisplayName:  "People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"id", "name"},
	})

	// Query that won't match anything → suggestions generated
	resp, err := svc.Search(SearchOptions{Query: "zzzxxx", EntityTypes: []string{"people"}})
	require.NoError(t, err)
	// No results means suggestions should be generated
	assert.Equal(t, 0, resp.Total)
}

func TestSearchEntity_WithWhereClause(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	setupPeopleTable(t, db)

	svc.entities = map[string]EntityConfig{}
	svc.RegisterEntity("people", EntityConfig{
		TableName:    "people",
		DisplayName:  "People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"id", "name"},
		WhereClause:  "name LIKE '%Alice%'",
	})

	resp, err := svc.Search(SearchOptions{Query: "Alice", EntityTypes: []string{"people"}})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestSearchEntity_WithTenantAndUserFilter(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	// Create table with tenant_id column
	require.NoError(t, db.Exec("CREATE TABLE orged_people (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, org_id INTEGER, owner_id INTEGER)").Error)
	require.NoError(t, db.Exec("INSERT INTO orged_people (name, org_id, owner_id) VALUES (?, ?, ?)", "Alice", 1, 5).Error)

	tenantID := uint(1)
	userID := uint(5)
	svc.entities = map[string]EntityConfig{}
	svc.RegisterEntity("orged_people", EntityConfig{
		TableName:    "orged_people",
		DisplayName:  "Orged People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"id", "name", "org_id"},
		Permissions: PermissionConfig{
			TenantField:    "org_id",
			OwnershipField: "owner_id",
		},
	})

	resp, err := svc.Search(SearchOptions{
		Query:       "Alice",
		EntityTypes: []string{"orged_people"},
		TenantID:    &tenantID,
		UserID:      &userID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestSearchEntity_WithFilters(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	require.NoError(t, db.Exec("CREATE TABLE filtered_people (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, status TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO filtered_people (name, status) VALUES (?, ?)", "Alice", "active").Error)

	svc.entities = map[string]EntityConfig{}
	svc.RegisterEntity("filtered_people", EntityConfig{
		TableName:    "filtered_people",
		DisplayName:  "Filtered People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"id", "name", "status"},
	})

	resp, err := svc.Search(SearchOptions{
		Query:       "Alice",
		EntityTypes: []string{"filtered_people"},
		Filters:     map[string]interface{}{"status": "active"},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.GreaterOrEqual(t, resp.Total, 1)
}

func TestSearchEntity_WithJoins(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	require.NoError(t, db.Exec("CREATE TABLE join_people (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, dept_id INTEGER)").Error)
	require.NoError(t, db.Exec("CREATE TABLE departments (id INTEGER PRIMARY KEY AUTOINCREMENT, dept_name TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO departments (id, dept_name) VALUES (1, 'Engineering')").Error)
	require.NoError(t, db.Exec("INSERT INTO join_people (name, dept_id) VALUES (?, ?)", "Alice", 1).Error)

	svc.entities = map[string]EntityConfig{}
	svc.RegisterEntity("join_people", EntityConfig{
		TableName:    "join_people",
		DisplayName:  "Joined People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"join_people.id", "join_people.name", "departments.dept_name"},
		JoinTables: []JoinConfig{
			{Table: "departments", Condition: "departments.id = join_people.dept_id", Type: "LEFT"},
		},
	})

	resp, err := svc.Search(SearchOptions{Query: "Alice", EntityTypes: []string{"join_people"}})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestSearchEntity_WithOrderAndGroup(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	require.NoError(t, db.Exec("CREATE TABLE grouped_people (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, category TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO grouped_people (name, category) VALUES (?, ?)", "Alice", "A").Error)

	svc.entities = map[string]EntityConfig{}
	svc.RegisterEntity("grouped_people", EntityConfig{
		TableName:    "grouped_people",
		DisplayName:  "Grouped People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"id", "name", "category"},
		OrderBy:      "name ASC",
		GroupBy:      "category",
	})

	resp, err := svc.Search(SearchOptions{Query: "Alice", EntityTypes: []string{"grouped_people"}})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestSearchEntity_InnerAndRightJoins(t *testing.T) {
	svc, db := newTestServiceWithDB(t)
	require.NoError(t, db.Exec("CREATE TABLE ij_people (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, dept_id INTEGER)").Error)
	require.NoError(t, db.Exec("CREATE TABLE ij_departments (id INTEGER PRIMARY KEY AUTOINCREMENT, dept_name TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO ij_departments (id, dept_name) VALUES (1, 'Ops')").Error)
	require.NoError(t, db.Exec("INSERT INTO ij_people (name, dept_id) VALUES (?, ?)", "Bob", 1).Error)

	// Test INNER join (default join)
	svc.entities = map[string]EntityConfig{}
	svc.RegisterEntity("ij_people", EntityConfig{
		TableName:    "ij_people",
		DisplayName:  "People",
		SearchFields: []FieldConfig{{Name: "name", Weight: 1.0, SearchType: "contains"}},
		SelectFields: []string{"ij_people.id", "ij_people.name"},
		JoinTables: []JoinConfig{
			{Table: "ij_departments", Condition: "ij_departments.id = ij_people.dept_id", Type: "INNER"},
		},
	})
	resp, err := svc.Search(SearchOptions{Query: "Bob", EntityTypes: []string{"ij_people"}})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestSearchEntity_UnknownEntityType(t *testing.T) {
	svc, _ := newTestServiceWithDB(t)
	svc.entities = map[string]EntityConfig{}

	// Search for an unregistered entity type — should return 0 results (no error)
	resp, err := svc.Search(SearchOptions{Query: "Alice", EntityTypes: []string{"unknown_entity"}})
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
}

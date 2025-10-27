package models

import (
	"fmt"
	"time"
)

// FuzzySearchLog represents search query logs for analytics
type FuzzySearchLog struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Query          string    `json:"query" gorm:"not null"`
	EntityTypes    string    `json:"entity_types"` // JSON array as string
	UserID         *uint     `json:"user_id,omitempty"`
	OrganizationID *uint     `json:"organization_id,omitempty"`
	ResultsCount   int       `json:"results_count"`
	ExecutionTime  int64     `json:"execution_time"` // microseconds
	IPAddress      string    `json:"ip_address,omitempty"`
	UserAgent      string    `json:"user_agent,omitempty"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SearchPreference represents user search preferences
type SearchPreference struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"not null"`
	EntityType     string    `json:"entity_type" gorm:"not null"`
	SortBy         string    `json:"sort_by"`
	SortOrder      string    `json:"sort_order"`
	DefaultFilters string    `json:"default_filters"` // JSON object as string
	Enabled        bool      `json:"enabled" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SavedSearch represents saved search queries
type SavedSearch struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	Name           string     `json:"name" gorm:"not null"`
	Description    string     `json:"description"`
	Query          string     `json:"query" gorm:"not null"`
	EntityTypes    string     `json:"entity_types"` // JSON array as string
	Filters        string     `json:"filters"`      // JSON object as string
	UserID         uint       `json:"user_id" gorm:"not null"`
	OrganizationID uint       `json:"organization_id" gorm:"not null"`
	IsPublic       bool       `json:"is_public" gorm:"default:false"`
	UseCount       int        `json:"use_count" gorm:"default:0"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// SearchableEntity represents entities that can be searched
type SearchableEntity struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	EntityType   string    `json:"entity_type" gorm:"unique;not null"`
	DisplayName  string    `json:"display_name" gorm:"not null"`
	Table        string    `json:"table" gorm:"not null"`
	SearchFields string    `json:"search_fields"` // JSON array of field configs
	SelectFields string    `json:"select_fields"` // JSON array
	JoinTables   string    `json:"join_tables"`   // JSON array of join configs
	WhereClause  string    `json:"where_clause"`
	OrderBy      string    `json:"order_by"`
	Permissions  string    `json:"permissions"` // JSON object
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SearchIndex represents search index entries for faster lookups
type SearchIndex struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	EntityType   string    `json:"entity_type" gorm:"not null;index"`
	EntityID     uint      `json:"entity_id" gorm:"not null;index"`
	FieldName    string    `json:"field_name" gorm:"not null"`
	FieldValue   string    `json:"field_value" gorm:"not null;index"`
	SearchVector string    `json:"search_vector"` // For full-text search
	Weight       float64   `json:"weight" gorm:"default:1.0"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SearchSynonym represents search synonyms for better matching
type SearchSynonym struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Term      string    `json:"term" gorm:"not null;index"`
	Synonyms  string    `json:"synonyms"` // JSON array of synonyms
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SearchStopword represents words to ignore in search
type SearchStopword struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Word      string    `json:"word" gorm:"unique;not null"`
	Language  string    `json:"language" gorm:"default:'en'"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Searchable interface defines methods that searchable entities should implement
type Searchable interface {
	GetSearchTitle() string
	GetSearchDescription() string
	GetSearchURL() string
	GetSearchData() map[string]interface{}
	GetEntityType() string
}

// SearchResultItem represents a single search result with metadata
type SearchResultItem struct {
	ID          interface{}            `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	URL         string                 `json:"url"`
	Score       float64                `json:"score"`
	Highlights  []string               `json:"highlights,omitempty"`
	Thumbnail   string                 `json:"thumbnail,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Data        map[string]interface{} `json:"data"`
	MetaData    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   *time.Time             `json:"created_at,omitempty"`
	UpdatedAt   *time.Time             `json:"updated_at,omitempty"`
}

// SearchFacet represents search facets for filtering
type SearchFacet struct {
	Name   string             `json:"name"`
	Values []SearchFacetValue `json:"values"`
}

// SearchFacetValue represents individual facet values
type SearchFacetValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// SearchAggregation represents search aggregation results
type SearchAggregation struct {
	Name    string                    `json:"name"`
	Type    string                    `json:"type"` // count, avg, sum, min, max
	Value   interface{}               `json:"value"`
	Buckets []SearchAggregationBucket `json:"buckets,omitempty"`
}

// SearchAggregationBucket represents aggregation bucket
type SearchAggregationBucket struct {
	Key   interface{} `json:"key"`
	Count int         `json:"count"`
	Value interface{} `json:"value,omitempty"`
}

// Advanced search request structures

// AdvancedSearchRequest represents an advanced search with complex filtering
type AdvancedSearchRequest struct {
	Query        string                     `json:"query"`
	Filters      []SearchFilter             `json:"filters,omitempty"`
	Sort         []SearchSort               `json:"sort,omitempty"`
	Facets       []string                   `json:"facets,omitempty"`
	Aggregations []SearchAggregationRequest `json:"aggregations,omitempty"`
	Highlight    SearchHighlight            `json:"highlight,omitempty"`
	Pagination   SearchPagination           `json:"pagination"`
}

// SearchFilter represents a search filter
type SearchFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, gte, lte, in, nin, like, exists
	Value    interface{} `json:"value"`
	Type     string      `json:"type,omitempty"` // string, number, date, boolean
}

// SearchSort represents sorting options
type SearchSort struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// SearchAggregationRequest represents an aggregation request
type SearchAggregationRequest struct {
	Name  string `json:"name"`
	Type  string `json:"type"` // terms, date_histogram, range, stats
	Field string `json:"field"`
	Size  int    `json:"size,omitempty"`
}

// SearchHighlight represents highlighting configuration
type SearchHighlight struct {
	Fields    []string `json:"fields"`
	PreTag    string   `json:"pre_tag,omitempty"`
	PostTag   string   `json:"post_tag,omitempty"`
	MaxLength int      `json:"max_length,omitempty"`
}

// SearchPagination represents pagination options
type SearchPagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// AdvancedSearchResponse represents advanced search results
type AdvancedSearchResponse struct {
	Query         string                 `json:"query"`
	Total         int                    `json:"total"`
	Results       []SearchResultItem     `json:"results"`
	Facets        []SearchFacet          `json:"facets,omitempty"`
	Aggregations  []SearchAggregation    `json:"aggregations,omitempty"`
	Suggestions   []string               `json:"suggestions,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time"`
	Debug         map[string]interface{} `json:"debug,omitempty"`
}

// Search analytics and reporting structures

// SearchAnalytics represents search analytics data
type SearchAnalytics struct {
	TotalSearches    int64                  `json:"total_searches"`
	PopularQueries   []PopularQuery         `json:"popular_queries"`
	NoResultQueries  []string               `json:"no_result_queries"`
	EntityTypeStats  map[string]int         `json:"entity_type_stats"`
	AvgExecutionTime float64                `json:"avg_execution_time"`
	PeakSearchHours  []int                  `json:"peak_search_hours"`
	UserSearchStats  map[string]interface{} `json:"user_search_stats"`
}

// PopularQuery represents popular search queries
type PopularQuery struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

// Utility methods for models

// TableName sets the table name for FuzzySearchLog
func (FuzzySearchLog) TableName() string {
	return "fuzzy_search_logs"
}

// TableName sets the table name for SearchPreference
func (SearchPreference) TableName() string {
	return "search_preferences"
}

// TableName sets the table name for SavedSearch
func (SavedSearch) TableName() string {
	return "saved_searches"
}

// TableName sets the table name for SearchableEntity
func (SearchableEntity) TableName() string {
	return "searchable_entities"
}

// TableName sets the table name for SearchIndex
func (SearchIndex) TableName() string {
	return "search_indexes"
}

// TableName sets the table name for SearchSynonym
func (SearchSynonym) TableName() string {
	return "search_synonyms"
}

// TableName sets the table name for SearchStopword
func (SearchStopword) TableName() string {
	return "search_stopwords"
}

// Implement Searchable interface for existing models

// User implements Searchable interface
func (u User) GetSearchTitle() string {
	return u.FirstName + " " + u.LastName
}

func (u User) GetSearchDescription() string {
	return "User: " + u.Email
}

func (u User) GetSearchURL() string {
	return "/api/v1/users/" + fmt.Sprintf("%d", u.ID)
}

func (u User) GetSearchData() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.ID,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"email":      u.Email,
		"username":   u.Username,
		"role":       u.Role,
		"active":     u.Active,
	}
}

func (u User) GetEntityType() string {
	return "users"
}

// Customer implements Searchable interface
func (c Customer) GetSearchTitle() string {
	return c.Name
}

func (c Customer) GetSearchDescription() string {
	desc := "Customer"
	if c.Email != "" {
		desc += ": " + c.Email
	}
	return desc
}

func (c Customer) GetSearchURL() string {
	return "/api/v1/customers/" + fmt.Sprintf("%d", c.ID)
}

func (c Customer) GetSearchData() map[string]interface{} {
	return map[string]interface{}{
		"id":      c.ID,
		"name":    c.Name,
		"email":   c.Email,
		"phone":   c.Phone,
		"street":  c.Street,
		"city":    c.City,
		"country": c.Country,
	}
}

func (c Customer) GetEntityType() string {
	return "customers"
}

// Contact implements Searchable interface
func (c Contact) GetSearchTitle() string {
	return c.FirstName + " " + c.LastName
}

func (c Contact) GetSearchDescription() string {
	return "Contact: " + c.Email
}

func (c Contact) GetSearchURL() string {
	return "/api/v1/contacts/" + fmt.Sprintf("%d", c.ID)
}

func (c Contact) GetSearchData() map[string]interface{} {
	return map[string]interface{}{
		"id":         c.ID,
		"first_name": c.FirstName,
		"last_name":  c.LastName,
		"email":      c.Email,
		"phone":      c.Phone,
		"notes":      c.Notes,
	}
}

func (c Contact) GetEntityType() string {
	return "contacts"
}

// Plan implements Searchable interface
func (p Plan) GetSearchTitle() string {
	return p.Name
}

func (p Plan) GetSearchDescription() string {
	return fmt.Sprintf("Plan: %s %.2f %s/%s", p.Name, p.Price, p.Currency, p.InvoicePeriod)
}

func (p Plan) GetSearchURL() string {
	return "/api/v1/plans/" + fmt.Sprintf("%d", p.ID)
}

func (p Plan) GetSearchData() map[string]interface{} {
	return map[string]interface{}{
		"id":             p.ID,
		"name":           p.Name,
		"description":    p.Description,
		"price":          p.Price,
		"currency":       p.Currency,
		"invoice_period": p.InvoicePeriod,
		"features":       p.Features,
		"active":         p.Active,
	}
}

func (p Plan) GetEntityType() string {
	return "plans"
}

// Email implements Searchable interface
func (e Email) GetSearchTitle() string {
	return e.Subject
}

func (e Email) GetSearchDescription() string {
	return fmt.Sprintf("Email to %s - %s", e.To, e.Status)
}

func (e Email) GetSearchURL() string {
	return "/api/v1/emails/" + fmt.Sprintf("%d", e.ID)
}

func (e Email) GetSearchData() map[string]interface{} {
	return map[string]interface{}{
		"id":      e.ID,
		"subject": e.Subject,
		"to":      e.To,
		"from":    e.From,
		"status":  e.Status,
		"sent_at": e.SentAt,
	}
}

func (e Email) GetEntityType() string {
	return "emails"
}

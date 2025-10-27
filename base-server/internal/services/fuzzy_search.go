package services

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FuzzySearchService provides generalized fuzzy search functionality
type FuzzySearchService struct {
	db       *gorm.DB
	config   FuzzySearchConfig
	entities map[string]EntityConfig
}

// FuzzySearchConfig holds configuration for fuzzy search behavior
type FuzzySearchConfig struct {
	MinSearchLength  int     `json:"min_search_length"`  // Minimum query length
	MaxResults       int     `json:"max_results"`        // Maximum results per entity type
	ScoreThreshold   float64 `json:"score_threshold"`    // Minimum relevance score
	EnableHighlight  bool    `json:"enable_highlight"`   // Highlight matching text
	CaseSensitive    bool    `json:"case_sensitive"`     // Case sensitive search
	ExactMatchBoost  float64 `json:"exact_match_boost"`  // Score boost for exact matches
	PrefixMatchBoost float64 `json:"prefix_match_boost"` // Score boost for prefix matches
	EnableStemming   bool    `json:"enable_stemming"`    // Enable word stemming
	EnableSynonyms   bool    `json:"enable_synonyms"`    // Enable synonym matching
}

// EntityConfig defines how to search within a specific entity type
type EntityConfig struct {
	TableName    string                 `json:"table_name"`
	DisplayName  string                 `json:"display_name"`
	SearchFields []FieldConfig          `json:"search_fields"`
	SelectFields []string               `json:"select_fields"`
	JoinTables   []JoinConfig           `json:"join_tables"`
	WhereClause  string                 `json:"where_clause"`
	OrderBy      string                 `json:"order_by"`
	GroupBy      string                 `json:"group_by"`
	Permissions  PermissionConfig       `json:"permissions"`
	MetaData     map[string]interface{} `json:"metadata"`
}

// FieldConfig defines how to search within a specific field
type FieldConfig struct {
	Name       string  `json:"name"`
	Weight     float64 `json:"weight"`      // Relevance weight for this field
	SearchType string  `json:"search_type"` // exact, prefix, contains, fuzzy, fulltext
	Boost      float64 `json:"boost"`       // Additional score boost
	Analyzer   string  `json:"analyzer"`    // Text analyzer type
	Required   bool    `json:"required"`    // Must match for result inclusion
	MinLength  int     `json:"min_length"`  // Minimum query length for this field
	Transform  string  `json:"transform"`   // Data transformation (lowercase, uppercase, etc.)
}

// JoinConfig defines table joins for enhanced search
type JoinConfig struct {
	Table     string `json:"table"`
	Condition string `json:"condition"`
	Type      string `json:"type"` // INNER, LEFT, RIGHT
	SearchIn  bool   `json:"search_in"`
}

// PermissionConfig defines access control for search
type PermissionConfig struct {
	RequireAuth    bool     `json:"require_auth"`
	AllowedRoles   []string `json:"allowed_roles"`
	OwnershipField string   `json:"ownership_field"`
	TenantField    string   `json:"organization_field"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID          interface{}            `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	URL         string                 `json:"url"`
	Score       float64                `json:"score"`
	Highlights  []string               `json:"highlights,omitempty"`
	Data        interface{}            `json:"data"`
	MetaData    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   *time.Time             `json:"created_at,omitempty"`
	UpdatedAt   *time.Time             `json:"updated_at,omitempty"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Query         string                 `json:"query"`
	Total         int                    `json:"total"`
	Results       []SearchResult         `json:"results"`
	Aggregations  map[string]interface{} `json:"aggregations,omitempty"`
	Suggestions   []string               `json:"suggestions,omitempty"`
	Categories    map[string]int         `json:"categories,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time"`
}

// SearchOptions defines search parameters
type SearchOptions struct {
	Query               string                 `json:"query"`
	EntityTypes         []string               `json:"entity_types,omitempty"`
	Filters             map[string]interface{} `json:"filters,omitempty"`
	SortBy              string                 `json:"sort_by,omitempty"`
	SortOrder           string                 `json:"sort_order,omitempty"`
	Offset              int                    `json:"offset"`
	Limit               int                    `json:"limit"`
	UserID              *uint                  `json:"user_id,omitempty"`
	TenantID            *uint                  `json:"tenant_id,omitempty"`
	IncludeCount        bool                   `json:"include_count"`
	IncludeAggregations bool                   `json:"include_aggregations"`
	HighlightFields     []string               `json:"highlight_fields,omitempty"`
}

// NewFuzzySearchService creates a new fuzzy search service
func NewFuzzySearchService(db *gorm.DB, config *FuzzySearchConfig) *FuzzySearchService {
	if config == nil {
		config = DefaultFuzzySearchConfig()
	}

	service := &FuzzySearchService{
		db:       db,
		config:   *config,
		entities: make(map[string]EntityConfig),
	}

	// Register default entities
	service.RegisterDefaultEntities()

	return service
}

// DefaultFuzzySearchConfig returns default search configuration
func DefaultFuzzySearchConfig() *FuzzySearchConfig {
	return &FuzzySearchConfig{
		MinSearchLength:  2,
		MaxResults:       50,
		ScoreThreshold:   0.1,
		EnableHighlight:  true,
		CaseSensitive:    false,
		ExactMatchBoost:  2.0,
		PrefixMatchBoost: 1.5,
		EnableStemming:   false,
		EnableSynonyms:   false,
	}
}

// RegisterEntity registers an entity type for searching
func (s *FuzzySearchService) RegisterEntity(entityType string, config EntityConfig) {
	s.entities[entityType] = config
}

// RegisterDefaultEntities registers common SaaS entities
func (s *FuzzySearchService) RegisterDefaultEntities() {
	// Users
	s.RegisterEntity("users", EntityConfig{
		TableName:   "users",
		DisplayName: "Users",
		SearchFields: []FieldConfig{
			{Name: "first_name", Weight: 1.0, SearchType: "contains"},
			{Name: "last_name", Weight: 1.0, SearchType: "contains"},
			{Name: "email", Weight: 0.8, SearchType: "contains"},
			{Name: "username", Weight: 0.9, SearchType: "prefix"},
		},
		SelectFields: []string{"id", "first_name", "last_name", "email", "username", "created_at"},
		WhereClause:  "active = true",
		OrderBy:      "first_name, last_name",
		Permissions: PermissionConfig{
			RequireAuth: true,
			TenantField: "organization_id",
		},
	})

	// Customers
	s.RegisterEntity("customers", EntityConfig{
		TableName:   "customers",
		DisplayName: "Customers",
		SearchFields: []FieldConfig{
			{Name: "name", Weight: 1.0, SearchType: "contains"},
			{Name: "email", Weight: 0.8, SearchType: "contains"},
			{Name: "phone", Weight: 0.6, SearchType: "contains"},
			{Name: "company", Weight: 0.7, SearchType: "contains"},
		},
		SelectFields: []string{"id", "name", "email", "phone", "company", "created_at"},
		OrderBy:      "name",
		Permissions: PermissionConfig{
			RequireAuth: true,
			TenantField: "organization_id",
		},
	})

	// Contacts
	s.RegisterEntity("contacts", EntityConfig{
		TableName:   "contacts",
		DisplayName: "Contacts",
		SearchFields: []FieldConfig{
			{Name: "name", Weight: 1.0, SearchType: "contains"},
			{Name: "email", Weight: 0.8, SearchType: "contains"},
			{Name: "phone", Weight: 0.6, SearchType: "contains"},
			{Name: "message", Weight: 0.4, SearchType: "fulltext"},
		},
		SelectFields: []string{"id", "name", "email", "phone", "subject", "created_at"},
		OrderBy:      "created_at DESC",
		Permissions: PermissionConfig{
			RequireAuth: true,
			TenantField: "organization_id",
		},
	})

	// Plans
	s.RegisterEntity("plans", EntityConfig{
		TableName:   "plans",
		DisplayName: "Plans",
		SearchFields: []FieldConfig{
			{Name: "name", Weight: 1.0, SearchType: "contains"},
			{Name: "description", Weight: 0.6, SearchType: "fulltext"},
			{Name: "features", Weight: 0.4, SearchType: "contains"},
		},
		SelectFields: []string{"id", "name", "description", "price", "currency", "billing_cycle"},
		WhereClause:  "active = true",
		OrderBy:      "price ASC",
		Permissions: PermissionConfig{
			RequireAuth: false, // Plans can be publicly searchable
		},
	})

	// Emails
	s.RegisterEntity("emails", EntityConfig{
		TableName:   "emails",
		DisplayName: "Emails",
		SearchFields: []FieldConfig{
			{Name: "subject", Weight: 1.0, SearchType: "contains"},
			{Name: "to_email", Weight: 0.8, SearchType: "contains"},
			{Name: "from_email", Weight: 0.6, SearchType: "contains"},
			{Name: "body", Weight: 0.4, SearchType: "fulltext"},
		},
		SelectFields: []string{"id", "subject", "to_email", "from_email", "status", "sent_at"},
		OrderBy:      "sent_at DESC",
		Permissions: PermissionConfig{
			RequireAuth: true,
			TenantField: "organization_id",
		},
	})
}

// Search performs fuzzy search across registered entities
func (s *FuzzySearchService) Search(options SearchOptions) (*SearchResponse, error) {
	startTime := time.Now()

	// Validate search query
	if len(strings.TrimSpace(options.Query)) < s.config.MinSearchLength {
		return &SearchResponse{
			Query:         options.Query,
			Total:         0,
			Results:       []SearchResult{},
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	// Determine entity types to search
	entityTypes := options.EntityTypes
	if len(entityTypes) == 0 {
		for entityType := range s.entities {
			entityTypes = append(entityTypes, entityType)
		}
	}

	// Apply default limits
	if options.Limit <= 0 || options.Limit > s.config.MaxResults {
		options.Limit = s.config.MaxResults
	}

	var allResults []SearchResult
	categories := make(map[string]int)

	// Search each entity type
	for _, entityType := range entityTypes {
		if config, exists := s.entities[entityType]; exists {
			results, err := s.searchEntity(entityType, config, options)
			if err != nil {
				continue // Skip entities with errors, don't fail entire search
			}

			allResults = append(allResults, results...)
			categories[entityType] = len(results)
		}
	}

	// Sort results by score
	allResults = s.sortResults(allResults, options)

	// Apply offset and limit
	total := len(allResults)
	if options.Offset >= total {
		allResults = []SearchResult{}
	} else {
		end := options.Offset + options.Limit
		if end > total {
			end = total
		}
		allResults = allResults[options.Offset:end]
	}

	// Generate suggestions if no results
	var suggestions []string
	if len(allResults) == 0 {
		suggestions = s.generateSuggestions(options.Query)
	}

	return &SearchResponse{
		Query:         options.Query,
		Total:         total,
		Results:       allResults,
		Categories:    categories,
		Suggestions:   suggestions,
		ExecutionTime: time.Since(startTime),
	}, nil
}

// searchEntity searches within a specific entity type
func (s *FuzzySearchService) searchEntity(entityType string, config EntityConfig, options SearchOptions) ([]SearchResult, error) {
	query := s.db.Table(config.TableName)

	// Apply joins
	for _, join := range config.JoinTables {
		switch strings.ToUpper(join.Type) {
		case "LEFT":
			query = query.Joins(fmt.Sprintf("LEFT JOIN %s ON %s", join.Table, join.Condition))
		case "RIGHT":
			query = query.Joins(fmt.Sprintf("RIGHT JOIN %s ON %s", join.Table, join.Condition))
		default:
			query = query.Joins(fmt.Sprintf("INNER JOIN %s ON %s", join.Table, join.Condition))
		}
	}

	// Apply base where clause
	if config.WhereClause != "" {
		query = query.Where(config.WhereClause)
	}

	// Apply organization filtering
	if options.TenantID != nil && config.Permissions.TenantField != "" {
		query = query.Where(config.Permissions.TenantField+" = ?", *options.TenantID)
	}

	// Apply user ownership filtering
	if options.UserID != nil && config.Permissions.OwnershipField != "" {
		query = query.Where(config.Permissions.OwnershipField+" = ?", *options.UserID)
	}

	// Apply additional filters
	for field, value := range options.Filters {
		query = query.Where(field+" = ?", value)
	}

	// Build search conditions
	searchConditions := s.buildSearchConditions(config.SearchFields, options.Query)
	if searchConditions != "" {
		query = query.Where(searchConditions)
	}

	// Apply select fields
	if len(config.SelectFields) > 0 {
		selectClause := strings.Join(config.SelectFields, ", ")
		query = query.Select(selectClause)
	}

	// Apply ordering
	if config.OrderBy != "" {
		query = query.Order(config.OrderBy)
	}

	// Apply group by
	if config.GroupBy != "" {
		query = query.Group(config.GroupBy)
	}

	// Execute query
	var rows []map[string]interface{}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	// Convert to SearchResult
	var results []SearchResult
	for _, row := range rows {
		result := s.convertToSearchResult(entityType, config, row, options)
		if result.Score >= s.config.ScoreThreshold {
			results = append(results, result)
		}
	}

	return results, nil
}

// buildSearchConditions creates SQL WHERE conditions for fuzzy search
func (s *FuzzySearchService) buildSearchConditions(fields []FieldConfig, searchQuery string) string {
	if searchQuery == "" {
		return ""
	}

	var conditions []string
	searchTerms := s.tokenizeQuery(searchQuery)

	for _, field := range fields {
		for _, term := range searchTerms {
			condition := s.buildFieldCondition(field, term)
			if condition != "" {
				conditions = append(conditions, condition)
			}
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	return "(" + strings.Join(conditions, " OR ") + ")"
}

// buildFieldCondition creates a condition for a specific field and search term
func (s *FuzzySearchService) buildFieldCondition(field FieldConfig, term string) string {
	if len(term) < field.MinLength {
		return ""
	}

	// Apply case sensitivity
	fieldName := field.Name
	searchTerm := term
	if !s.config.CaseSensitive {
		fieldName = fmt.Sprintf("LOWER(%s)", field.Name)
		searchTerm = strings.ToLower(term)
	}

	switch field.SearchType {
	case "exact":
		return fmt.Sprintf("%s = '%s'", fieldName, searchTerm)
	case "prefix":
		return fmt.Sprintf("%s LIKE '%s%%'", fieldName, searchTerm)
	case "contains":
		return fmt.Sprintf("%s LIKE '%%%s%%'", fieldName, searchTerm)
	case "fulltext":
		// PostgreSQL full-text search
		return fmt.Sprintf("to_tsvector('english', %s) @@ plainto_tsquery('english', '%s')", field.Name, searchTerm)
	case "fuzzy":
		// Simple fuzzy matching using SIMILAR TO (PostgreSQL)
		fuzzyPattern := s.buildFuzzyPattern(searchTerm)
		return fmt.Sprintf("%s SIMILAR TO '%s'", fieldName, fuzzyPattern)
	default:
		return fmt.Sprintf("%s LIKE '%%%s%%'", fieldName, searchTerm)
	}
}

// buildFuzzyPattern creates a fuzzy search pattern
func (s *FuzzySearchService) buildFuzzyPattern(term string) string {
	// Simple fuzzy pattern: allow single character substitutions
	if len(term) <= 3 {
		return term
	}

	// For longer terms, create pattern with optional characters
	pattern := ""
	for i, char := range term {
		if i > 0 {
			pattern += ".*?"
		}
		pattern += string(char)
	}
	return pattern
}

// tokenizeQuery splits search query into individual terms
func (s *FuzzySearchService) tokenizeQuery(query string) []string {
	// Simple tokenization - can be enhanced with stemming, stop words, etc.
	terms := strings.Fields(strings.TrimSpace(query))

	// Remove empty terms and apply minimum length
	var validTerms []string
	for _, term := range terms {
		if len(term) >= s.config.MinSearchLength {
			validTerms = append(validTerms, term)
		}
	}

	return validTerms
}

// convertToSearchResult converts database row to SearchResult
func (s *FuzzySearchService) convertToSearchResult(entityType string, config EntityConfig, row map[string]interface{}, options SearchOptions) SearchResult {
	result := SearchResult{
		Type:     entityType,
		Data:     row,
		MetaData: config.MetaData,
	}

	// Extract ID
	if id, exists := row["id"]; exists {
		result.ID = id
	}

	// Build title and description based on entity type
	result.Title, result.Description = s.buildTitleDescription(entityType, row)

	// Generate URL
	result.URL = s.buildURL(entityType, result.ID)

	// Calculate relevance score
	result.Score = s.calculateRelevanceScore(config.SearchFields, row, options.Query)

	// Add highlights if enabled
	if s.config.EnableHighlight {
		result.Highlights = s.generateHighlights(config.SearchFields, row, options.Query)
	}

	// Extract timestamps
	if createdAt, exists := row["created_at"]; exists {
		if t, ok := createdAt.(time.Time); ok {
			result.CreatedAt = &t
		}
	}
	if updatedAt, exists := row["updated_at"]; exists {
		if t, ok := updatedAt.(time.Time); ok {
			result.UpdatedAt = &t
		}
	}

	return result
}

// buildTitleDescription creates title and description for search results
func (s *FuzzySearchService) buildTitleDescription(entityType string, row map[string]interface{}) (string, string) {
	switch entityType {
	case "users":
		title := fmt.Sprintf("%v %v", getField(row, "first_name"), getField(row, "last_name"))
		description := fmt.Sprintf("Email: %v", getField(row, "email"))
		return strings.TrimSpace(title), description

	case "customers":
		title := fmt.Sprintf("%v", getField(row, "name"))
		description := fmt.Sprintf("Email: %v", getField(row, "email"))
		if company := getField(row, "company"); company != "" {
			description += fmt.Sprintf(" | Company: %v", company)
		}
		return title, description

	case "contacts":
		title := fmt.Sprintf("%v", getField(row, "name"))
		description := fmt.Sprintf("Subject: %v", getField(row, "subject"))
		return title, description

	case "plans":
		title := fmt.Sprintf("%v", getField(row, "name"))
		price := getField(row, "price")
		currency := getField(row, "currency")
		description := fmt.Sprintf("Price: %v %v", price, currency)
		return title, description

	case "emails":
		title := fmt.Sprintf("%v", getField(row, "subject"))
		description := fmt.Sprintf("To: %v | Status: %v", getField(row, "to_email"), getField(row, "status"))
		return title, description

	default:
		// Generic fallback
		if name := getField(row, "name"); name != "" {
			return name, fmt.Sprintf("%s record", entityType)
		}
		if title := getField(row, "title"); title != "" {
			return title, fmt.Sprintf("%s record", entityType)
		}
		return fmt.Sprintf("%s #%v", entityType, getField(row, "id")), ""
	}
}

// buildURL generates a URL for the search result
func (s *FuzzySearchService) buildURL(entityType string, id interface{}) string {
	if id == nil {
		return ""
	}
	return fmt.Sprintf("/api/v1/%s/%v", entityType, id)
}

// calculateRelevanceScore calculates relevance score for a search result
func (s *FuzzySearchService) calculateRelevanceScore(fields []FieldConfig, row map[string]interface{}, query string) float64 {
	if query == "" {
		return 0
	}

	searchTerms := s.tokenizeQuery(query)
	totalScore := 0.0

	for _, field := range fields {
		fieldValue := fmt.Sprintf("%v", getField(row, field.Name))
		if fieldValue == "" {
			continue
		}

		fieldScore := s.calculateFieldScore(field, fieldValue, searchTerms)
		totalScore += fieldScore * field.Weight
	}

	return totalScore
}

// calculateFieldScore calculates score for a specific field
func (s *FuzzySearchService) calculateFieldScore(field FieldConfig, fieldValue string, searchTerms []string) float64 {
	fieldValue = strings.ToLower(fieldValue)
	maxScore := 0.0

	for _, term := range searchTerms {
		termLower := strings.ToLower(term)
		score := 0.0

		// Exact match
		if fieldValue == termLower {
			score = s.config.ExactMatchBoost
		} else if strings.HasPrefix(fieldValue, termLower) {
			// Prefix match
			score = s.config.PrefixMatchBoost
		} else if strings.Contains(fieldValue, termLower) {
			// Contains match
			score = 1.0
		} else {
			// Fuzzy match (simple Levenshtein-like scoring)
			score = s.calculateFuzzyScore(fieldValue, termLower)
		}

		// Apply field boost
		score *= (1.0 + field.Boost)

		if score > maxScore {
			maxScore = score
		}
	}

	return maxScore
}

// calculateFuzzyScore calculates fuzzy match score
func (s *FuzzySearchService) calculateFuzzyScore(text, term string) float64 {
	if len(term) < 3 {
		return 0
	}

	// Simple character overlap ratio
	matches := 0
	for _, char := range term {
		if strings.ContainsRune(text, char) {
			matches++
		}
	}

	ratio := float64(matches) / float64(len(term))
	if ratio < 0.5 {
		return 0
	}

	return ratio * 0.5 // Lower score for fuzzy matches
}

// generateHighlights generates highlighted text snippets
func (s *FuzzySearchService) generateHighlights(fields []FieldConfig, row map[string]interface{}, query string) []string {
	var highlights []string
	searchTerms := s.tokenizeQuery(query)

	for _, field := range fields {
		fieldValue := fmt.Sprintf("%v", getField(row, field.Name))
		if fieldValue == "" {
			continue
		}

		highlighted := s.highlightText(fieldValue, searchTerms)
		if highlighted != fieldValue {
			highlights = append(highlights, highlighted)
		}
	}

	return highlights
}

// highlightText highlights search terms in text
func (s *FuzzySearchService) highlightText(text string, terms []string) string {
	highlighted := text

	for _, term := range terms {
		if !s.config.CaseSensitive {
			// Case-insensitive highlighting - find all occurrences
			lowerText := strings.ToLower(highlighted)
			lowerTerm := strings.ToLower(term)
			if strings.Contains(lowerText, lowerTerm) {
				highlighted = strings.ReplaceAll(highlighted, term, fmt.Sprintf("<mark>%s</mark>", term))
			}
		} else {
			highlighted = strings.ReplaceAll(highlighted, term, fmt.Sprintf("<mark>%s</mark>", term))
		}
	}

	return highlighted
}

// sortResults sorts search results by relevance score
func (s *FuzzySearchService) sortResults(results []SearchResult, options SearchOptions) []SearchResult {
	// Default sort by score descending
	if options.SortBy == "" || options.SortBy == "score" {
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].Score < results[j].Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
		return results
	}

	// Custom sorting (can be enhanced)
	return results
}

// generateSuggestions generates search suggestions when no results found
func (s *FuzzySearchService) generateSuggestions(query string) []string {
	// Simple suggestion generation - can be enhanced with ML/AI
	suggestions := []string{
		"Try using different keywords",
		"Check your spelling",
		"Use more general terms",
		"Try searching in specific categories",
	}

	// Add entity type suggestions
	for _, config := range s.entities {
		suggestions = append(suggestions, fmt.Sprintf("Search in %s", config.DisplayName))
	}

	return suggestions
}

// GetEntityTypes returns all registered entity types
func (s *FuzzySearchService) GetEntityTypes() map[string]EntityConfig {
	return s.entities
}

// GetConfig returns the current search configuration
func (s *FuzzySearchService) GetConfig() FuzzySearchConfig {
	return s.config
}

// UpdateConfig updates the search configuration
func (s *FuzzySearchService) UpdateConfig(config FuzzySearchConfig) {
	s.config = config
}

// Helper function to safely get field value from map
func getField(row map[string]interface{}, field string) string {
	if value, exists := row[field]; exists && value != nil {
		return fmt.Sprintf("%v", value)
	}
	return ""
}

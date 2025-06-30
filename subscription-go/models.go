package subscription

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

type Subscription struct {
	BaseModel
	AccountID            uuid.UUID          `gorm:"not null" json:"account_id"`
	PlanID               uuid.UUID          `gorm:"not null" json:"plan_id"`
	Plan                 SubscriptionPlan   `json:"plan,omitempty"`
	Status               SubscriptionStatus `gorm:"not null" json:"status"`
	CurrentPeriodStart   time.Time          `json:"current_period_start"`
	CurrentPeriodEnd     time.Time          `json:"current_period_end"`
	CancelAt             *time.Time         `json:"cancel_at"`
	CancelledAt          *time.Time         `json:"cancelled_at"`
	StripeCustomerID     string             `json:"-"`
	StripeSubscriptionID string             `json:"-"`
}

type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "active"
	SubscriptionPending  SubscriptionStatus = "pending"
	SubscriptionCanceled SubscriptionStatus = "canceled"
	SubscriptionExpired  SubscriptionStatus = "expired"
)

type SubscriptionPlan struct {
	BaseModel
	Name          string       `gorm:"not null" json:"name"`
	Code          string       `gorm:"uniqueIndex;not null" json:"code"`
	Description   string       `json:"description"`
	Price         float64      `gorm:"not null" json:"price"`
	Currency      string       `gorm:"default:'USD'" json:"currency"`
	Interval      string       `gorm:"default:'month'" json:"interval"`
	Features      PlanFeatures `gorm:"type:jsonb" json:"features"`
	IsActive      bool         `gorm:"default:true" json:"is_active"`
	IsVisible     bool         `gorm:"default:true" json:"is_visible"`
	TrialDays     int          `gorm:"default:0" json:"trial_days"`
	StripePriceID string       `json:"-"`
}

// PlanFeatures uses generic maps for flexibility
type PlanFeatures struct {
	Limits map[string]int64       `json:"limits"`
	Flags  map[string]bool        `json:"flags"`
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// Helper methods for type-safe access
func (pf PlanFeatures) GetLimit(key string) int64 {
	if pf.Limits == nil {
		return 0
	}
	return pf.Limits[key]
}

func (pf PlanFeatures) GetFlag(key string) bool {
	if pf.Flags == nil {
		return false
	}
	return pf.Flags[key]
}

func (pf PlanFeatures) IsUnlimited(key string) bool {
	return pf.GetLimit(key) == -1
}


// GORM Scanner/Valuer interfaces for JSONB
func (pf PlanFeatures) Value() (driver.Value, error) {
	return json.Marshal(pf)
}

func (pf *PlanFeatures) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte("{}"), pf)
	}
	return json.Unmarshal(bytes, pf)
}

func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionActive && time.Now().Before(s.CurrentPeriodEnd)
}

func (s *Subscription) CanAddResource(resourceType string, currentCount int, limitKey string) bool {
	if limitKey == "" {
		return true
	}

	limit := s.Plan.Features.GetLimit(limitKey)
	if limit == -1 {
		return true
	}
	return int64(currentCount) < limit
}

// SubscriptionUsage tracks usage metrics for a subscription billing period
type SubscriptionUsage struct {
	BaseModel
	SubscriptionID   uuid.UUID    `gorm:"not null;index" json:"subscription_id"`
	Subscription     Subscription `json:"subscription,omitempty"`
	PeriodStart      time.Time    `gorm:"not null;index" json:"period_start"`
	PeriodEnd        time.Time    `gorm:"not null;index" json:"period_end"`
	FeedbacksCount   int          `gorm:"default:0" json:"feedbacks_count"`
	RestaurantsCount int          `gorm:"default:0" json:"restaurants_count"`
	LocationsCount   int          `gorm:"default:0" json:"locations_count"`
	QRCodesCount     int          `gorm:"default:0" json:"qr_codes_count"`
	TeamMembersCount int          `gorm:"default:0" json:"team_members_count"`
	LastUpdatedAt    time.Time    `json:"last_updated_at"`
}

// UsageEvent tracks individual usage events for auditing
type UsageEvent struct {
	BaseModel
	SubscriptionID uuid.UUID `gorm:"not null;index" json:"subscription_id"`
	EventType      string    `gorm:"not null" json:"event_type"`
	ResourceType   string    `gorm:"not null" json:"resource_type"`
	ResourceID     uuid.UUID `json:"resource_id"`
	Metadata       string    `gorm:"type:jsonb" json:"metadata"`
}

// Event types
const (
	EventTypeCreate = "create"
	EventTypeDelete = "delete"
	EventTypeUpdate = "update"
)

// Resource types - these can be customized per application
const (
	ResourceTypeFeedback   = "feedback"
	ResourceTypeRestaurant = "restaurant"
	ResourceTypeLocation   = "location"
	ResourceTypeQRCode     = "qr_code"
	ResourceTypeTeamMember = "team_member"
)

// CanAddResource checks if a resource can be added based on plan limits
func (u *SubscriptionUsage) CanAddResource(resourceType string, plan PlanFeatures, limitKey string) (bool, string) {
	if limitKey == "" {
		return true, ""
	}

	var currentUsage int64
	switch resourceType {
	case ResourceTypeFeedback:
		currentUsage = int64(u.FeedbacksCount)
	case ResourceTypeRestaurant:
		currentUsage = int64(u.RestaurantsCount)
	case ResourceTypeLocation:
		currentUsage = int64(u.LocationsCount)
	case ResourceTypeQRCode:
		currentUsage = int64(u.QRCodesCount)
	case ResourceTypeTeamMember:
		currentUsage = int64(u.TeamMembersCount)
	default:
		return true, ""
	}

	limit := plan.GetLimit(limitKey)
	if limit == -1 {
		return true, ""
	}

	if currentUsage >= limit {
		def, _ := GetFeatureDefinition(limitKey)
		displayName := limitKey
		if def.DisplayName != "" {
			displayName = def.DisplayName
		}
		return false, fmt.Sprintf("%s limit reached", displayName)
	}

	return true, ""
}

// FeatureType represents the type of feature
type FeatureType string

const (
	FeatureTypeLimit  FeatureType = "limit"
	FeatureTypeFlag   FeatureType = "flag"
	FeatureTypeCustom FeatureType = "custom"
)

// FeatureDefinition defines metadata for a feature
type FeatureDefinition struct {
	Key           string                 `json:"key"`
	Type          FeatureType            `json:"type"`
	DisplayName   string                 `json:"display_name"`
	Description   string                 `json:"description"`
	Unit          string                 `json:"unit,omitempty"`
	UnlimitedText string                 `json:"unlimited_text,omitempty"`
	Format        string                 `json:"format,omitempty"`
	Icon          string                 `json:"icon,omitempty"`
	Category      string                 `json:"category,omitempty"`
	SortOrder     int                    `json:"sort_order"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// FeatureRegistry is now configurable and should be injected via service initialization
var FeatureRegistry map[string]FeatureDefinition

// InitializeFeatureRegistry sets up the feature registry with provided definitions
func InitializeFeatureRegistry(features map[string]FeatureDefinition) {
	FeatureRegistry = features
}

// GetAllFeatures returns all features in the registry
func GetAllFeatures() map[string]FeatureDefinition {
	if FeatureRegistry == nil {
		return make(map[string]FeatureDefinition)
	}
	return FeatureRegistry
}

// GetFeatureDefinition returns the definition for a feature key
func GetFeatureDefinition(key string) (FeatureDefinition, bool) {
	if FeatureRegistry == nil {
		return FeatureDefinition{}, false
	}
	def, exists := FeatureRegistry[key]
	return def, exists
}

// GetFeaturesByCategory returns all features in a category
func GetFeaturesByCategory(category string) []FeatureDefinition {
	var features []FeatureDefinition
	if FeatureRegistry == nil {
		return features
	}
	for _, def := range FeatureRegistry {
		if def.Category == category {
			features = append(features, def)
		}
	}
	return features
}

// FormatFeatureValue formats a feature value for display
func FormatFeatureValue(key string, value interface{}) string {
	def, exists := FeatureRegistry[key]
	if !exists {
		return ""
	}

	switch def.Type {
	case FeatureTypeLimit:
		limitValue, ok := value.(int64)
		if !ok {
			if intVal, ok := value.(int); ok {
				limitValue = int64(intVal)
			} else {
				return ""
			}
		}

		if limitValue == -1 {
			return def.UnlimitedText
		}

		// Simple format replacement
		result := def.Format
		if result == "" {
			result = "{value} {unit}"
		}
		result = replaceValue(result, "{value}", limitValue)
		result = replaceValue(result, "{unit}", def.Unit)
		return result

	case FeatureTypeFlag:
		if boolVal, ok := value.(bool); ok && boolVal {
			return def.DisplayName
		}
		return ""

	default:
		return ""
	}
}

func replaceValue(s string, old string, value interface{}) string {
	return strings.ReplaceAll(s, old, fmt.Sprintf("%v", value))
}
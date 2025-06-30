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

// Common limit keys as constants for type safety
const (
	LimitRestaurants            = "max_restaurants"
	LimitLocationsPerRestaurant = "max_locations_per_restaurant"
	LimitQRCodesPerLocation     = "max_qr_codes_per_location"
	LimitFeedbacksPerMonth      = "max_feedbacks_per_month"
	LimitTeamMembers            = "max_team_members"
	LimitStorageGB              = "max_storage_gb"
	LimitAPICallsPerHour        = "max_api_calls_per_hour"
)

// Common feature flags
const (
	FlagAdvancedAnalytics = "advanced_analytics"
	FlagCustomBranding    = "custom_branding"
	FlagAPIAccess         = "api_access"
	FlagPrioritySupport   = "priority_support"
	FlagWhiteLabel        = "white_label"
	FlagCustomDomain      = "custom_domain"
)

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

func (s *Subscription) CanAddResource(resourceType string, currentCount int) bool {
	var limitKey string
	switch resourceType {
	case "restaurant":
		limitKey = LimitRestaurants
	case "team_member":
		limitKey = LimitTeamMembers
	case "feedback":
		limitKey = LimitFeedbacksPerMonth
	default:
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

// Resource types
const (
	ResourceTypeFeedback   = "feedback"
	ResourceTypeRestaurant = "restaurant"
	ResourceTypeLocation   = "location"
	ResourceTypeQRCode     = "qr_code"
	ResourceTypeTeamMember = "team_member"
)

// CanAddResource checks if a resource can be added based on plan limits
func (u *SubscriptionUsage) CanAddResource(resourceType string, plan PlanFeatures) (bool, string) {
	var limitKey string
	var currentUsage int64

	switch resourceType {
	case ResourceTypeFeedback:
		limitKey = LimitFeedbacksPerMonth
		currentUsage = int64(u.FeedbacksCount)
	case ResourceTypeRestaurant:
		limitKey = LimitRestaurants
		currentUsage = int64(u.RestaurantsCount)
	case ResourceTypeLocation:
		limitKey = LimitLocationsPerRestaurant
		currentUsage = int64(u.LocationsCount)
	case ResourceTypeQRCode:
		limitKey = LimitQRCodesPerLocation
		currentUsage = int64(u.QRCodesCount)
	case ResourceTypeTeamMember:
		limitKey = LimitTeamMembers
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
		return false, fmt.Sprintf("%s limit reached", def.DisplayName)
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

// FeatureRegistry holds all feature definitions
var FeatureRegistry = map[string]FeatureDefinition{
	// Limits
	LimitRestaurants: {
		Key:           LimitRestaurants,
		Type:          FeatureTypeLimit,
		DisplayName:   "Restaurants",
		Description:   "Maximum number of restaurants",
		Unit:          "restaurants",
		UnlimitedText: "Unlimited restaurants",
		Format:        "{value} restaurant(s)",
		Icon:          "store",
		Category:      "core",
		SortOrder:     1,
	},
	LimitFeedbacksPerMonth: {
		Key:           LimitFeedbacksPerMonth,
		Type:          FeatureTypeLimit,
		DisplayName:   "Monthly Feedbacks",
		Description:   "Maximum feedbacks per month",
		Unit:          "feedbacks/month",
		UnlimitedText: "Unlimited feedbacks",
		Format:        "{value} feedbacks/month",
		Icon:          "message-square",
		Category:      "core",
		SortOrder:     2,
	},
	LimitQRCodesPerLocation: {
		Key:           LimitQRCodesPerLocation,
		Type:          FeatureTypeLimit,
		DisplayName:   "QR Codes",
		Description:   "QR codes per location",
		Unit:          "QR codes/location",
		UnlimitedText: "Unlimited QR codes",
		Format:        "{value} QR codes per location",
		Icon:          "qr-code",
		Category:      "core",
		SortOrder:     3,
	},
	LimitTeamMembers: {
		Key:           LimitTeamMembers,
		Type:          FeatureTypeLimit,
		DisplayName:   "Team Members",
		Description:   "Maximum team members",
		Unit:          "members",
		UnlimitedText: "Unlimited team members",
		Format:        "{value} team member(s)",
		Icon:          "users",
		Category:      "collaboration",
		SortOrder:     4,
	},
	LimitStorageGB: {
		Key:           LimitStorageGB,
		Type:          FeatureTypeLimit,
		DisplayName:   "Storage",
		Description:   "Storage space for media files",
		Unit:          "GB",
		UnlimitedText: "Unlimited storage",
		Format:        "{value} GB storage",
		Icon:          "hard-drive",
		Category:      "resources",
		SortOrder:     5,
	},
	LimitAPICallsPerHour: {
		Key:           LimitAPICallsPerHour,
		Type:          FeatureTypeLimit,
		DisplayName:   "API Rate Limit",
		Description:   "API calls per hour",
		Unit:          "calls/hour",
		UnlimitedText: "Unlimited API calls",
		Format:        "{value} API calls/hour",
		Icon:          "activity",
		Category:      "developer",
		SortOrder:     10,
	},

	// Flags
	FlagAdvancedAnalytics: {
		Key:         FlagAdvancedAnalytics,
		Type:        FeatureTypeFlag,
		DisplayName: "Advanced Analytics",
		Description: "Detailed insights and reporting",
		Icon:        "bar-chart",
		Category:    "analytics",
		SortOrder:   20,
	},
	FlagCustomBranding: {
		Key:         FlagCustomBranding,
		Type:        FeatureTypeFlag,
		DisplayName: "Custom Branding",
		Description: "Customize with your brand",
		Icon:        "palette",
		Category:    "customization",
		SortOrder:   21,
	},
	FlagAPIAccess: {
		Key:         FlagAPIAccess,
		Type:        FeatureTypeFlag,
		DisplayName: "API Access",
		Description: "Programmatic access via API",
		Icon:        "code",
		Category:    "developer",
		SortOrder:   22,
	},
	FlagPrioritySupport: {
		Key:         FlagPrioritySupport,
		Type:        FeatureTypeFlag,
		DisplayName: "Priority Support",
		Description: "24/7 priority customer support",
		Icon:        "headphones",
		Category:    "support",
		SortOrder:   23,
	},
	FlagWhiteLabel: {
		Key:         FlagWhiteLabel,
		Type:        FeatureTypeFlag,
		DisplayName: "White Label",
		Description: "Remove branding",
		Icon:        "eye-off",
		Category:    "customization",
		SortOrder:   24,
	},
	FlagCustomDomain: {
		Key:         FlagCustomDomain,
		Type:        FeatureTypeFlag,
		DisplayName: "Custom Domain",
		Description: "Use your own domain",
		Icon:        "globe",
		Category:    "customization",
		SortOrder:   25,
	},
}

// GetFeatureDefinition returns the definition for a feature key
func GetFeatureDefinition(key string) (FeatureDefinition, bool) {
	def, exists := FeatureRegistry[key]
	return def, exists
}

// GetFeaturesByCategory returns all features in a category
func GetFeaturesByCategory(category string) []FeatureDefinition {
	var features []FeatureDefinition
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
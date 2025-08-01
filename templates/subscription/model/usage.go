package subscriptionmodel

import (
	"time"
	
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Usage represents resource usage
type Usage struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AccountID uuid.UUID      `json:"account_id" gorm:"type:uuid;not null;index"`
	Resource  string         `json:"resource" gorm:"not null;index"`
	Quantity  int64          `json:"quantity" gorm:"not null;default:0"`
	Period    time.Time      `json:"period" gorm:"not null;index"` // Month start
	Metadata  datatypes.JSON `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// GetID returns the usage ID
func (u *Usage) GetID() uuid.UUID {
	return u.ID
}

// GetAccountID returns the account ID
func (u *Usage) GetAccountID() uuid.UUID {
	return u.AccountID
}

// GetResource returns the resource name
func (u *Usage) GetResource() string {
	return u.Resource
}

// GetQuantity returns the quantity used
func (u *Usage) GetQuantity() int64 {
	return u.Quantity
}

// GetPeriod returns the period
func (u *Usage) GetPeriod() time.Time {
	return u.Period
}

// GetMetadata returns the metadata
func (u *Usage) GetMetadata() map[string]interface{} {
	var metadata map[string]interface{}
	if u.Metadata != nil {
		u.Metadata.Scan(&metadata)
	}
	return metadata
}

// GetCreatedAt returns creation time
func (u *Usage) GetCreatedAt() time.Time {
	return u.CreatedAt
}

// IncrementQuantity increments the quantity
func (u *Usage) IncrementQuantity(amount int64) {
	u.Quantity += amount
}
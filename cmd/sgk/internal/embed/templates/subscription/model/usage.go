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

func (u *Usage) GetID() uuid.UUID {
	return u.ID
}

func (u *Usage) GetAccountID() uuid.UUID {
	return u.AccountID
}

func (u *Usage) GetResource() string {
	return u.Resource
}

func (u *Usage) GetQuantity() int64 {
	return u.Quantity
}

func (u *Usage) GetPeriod() time.Time {
	return u.Period
}

func (u *Usage) GetMetadata() map[string]any {
	var metadata map[string]any
	if u.Metadata != nil {
		u.Metadata.Scan(&metadata)
	}
	return metadata
}

func (u *Usage) GetCreatedAt() time.Time {
	return u.CreatedAt
}

func (u *Usage) IncrementQuantity(amount int64) {
	u.Quantity += amount
}

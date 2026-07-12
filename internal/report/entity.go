package report

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WasteSummaryRow struct {
	Type   string `db:"type" json:"type"`
	Status string `db:"status" json:"status"`
	Count  int64  `db:"count" json:"count"`
}

type PaymentStatusRow struct {
	Status      string          `db:"status" json:"status"`
	Count       int64           `db:"count" json:"count"`
	TotalAmount decimal.Decimal `db:"total_amount" json:"total_amount"`
}

type PaymentSummary struct {
	ByStatus     []PaymentStatusRow `json:"by_status"`
	TotalRevenue decimal.Decimal    `json:"total_revenue"`
}

type HouseholdInfo struct {
	ID        uuid.UUID `db:"id" json:"id"`
	OwnerName string    `db:"owner_name" json:"owner_name"`
	Address   string    `db:"address" json:"address"`
}

type PickupRow struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	Type       string     `db:"type" json:"type"`
	Status     string     `db:"status" json:"status"`
	PickupDate *time.Time `db:"pickup_date" json:"pickup_date"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

type PaymentRow struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	WasteID     uuid.UUID       `db:"waste_id" json:"waste_id"`
	Amount      decimal.Decimal `db:"amount" json:"amount"`
	Status      string          `db:"status" json:"status"`
	PaymentDate *time.Time      `db:"payment_date" json:"payment_date"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}

type HouseholdHistory struct {
	Household HouseholdInfo `json:"household"`
	Pickups   []PickupRow   `json:"pickups"`
	Payments  []PaymentRow  `json:"payments"`
}

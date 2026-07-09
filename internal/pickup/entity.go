package pickup

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/httpres"
)

type WasteType string

const (
	TypeOrganic    WasteType = "organic"
	TypePlastic    WasteType = "plastic"
	TypePaper      WasteType = "paper"
	TypeElectronic WasteType = "electronic"
)

func (t WasteType) Valid() bool {
	switch t {
	case TypeOrganic, TypePlastic, TypePaper, TypeElectronic:
		return true
	}
	return false
}

type Status string

const (
	StatusPending   Status = "pending"
	StatusScheduled Status = "scheduled"
	StatusCompleted Status = "completed"
	StatusCanceled  Status = "canceled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusScheduled, StatusCompleted, StatusCanceled:
		return true
	}
	return false
}

type Pickup struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	HouseholdID uuid.UUID  `db:"household_id" json:"household_id"`
	Type        WasteType  `db:"type" json:"type"`
	Status      Status     `db:"status" json:"status"`
	PickupDate  *time.Time `db:"pickup_date" json:"pickup_date"`
	SafetyCheck *bool      `db:"safety_check" json:"safety_check"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type CreateRequest struct {
	HouseholdID uuid.UUID `json:"household_id"`
	Type        WasteType `json:"type"`
	SafetyCheck *bool     `json:"safety_check"`
}

func (r CreateRequest) Validate() error {
	if r.HouseholdID == uuid.Nil {
		return apperr.New(http.StatusBadRequest, "household_id is required")
	}
	if !r.Type.Valid() {
		return apperr.New(http.StatusBadRequest, "type must be organic, plastic, paper or electronic")
	}
	if r.Type == TypeElectronic && r.SafetyCheck == nil {
		return apperr.New(http.StatusBadRequest, "safety_check is required for electronic pickups")
	}
	return nil
}

type ListFilter struct {
	Status      string
	HouseholdID uuid.UUID
	Page        httpres.Page
}

func ParseListFilter(r *http.Request) (ListFilter, error) {
	f := ListFilter{Page: httpres.ParsePage(r)}
	if s := r.URL.Query().Get("status"); s != "" {
		if !Status(s).Valid() {
			return f, apperr.New(http.StatusBadRequest, "invalid status filter")
		}
		f.Status = s
	}
	if h := r.URL.Query().Get("household_id"); h != "" {
		id, err := uuid.Parse(h)
		if err != nil {
			return f, apperr.New(http.StatusBadRequest, "invalid household_id filter")
		}
		f.HouseholdID = id
	}
	return f, nil
}

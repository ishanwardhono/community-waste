package payment

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
	"github.com/ishanwardhono/community-waste/pkg/httpres"
)

const maxProofSize = 5 << 20

type ProofFile struct {
	Name        string
	Size        int64
	ContentType string
	Reader      io.Reader
}

func (p ProofFile) Validate() error {
	if p.Size <= 0 || p.Size > maxProofSize {
		return apperr.New(http.StatusBadRequest, "proof file must be present and under 5MB")
	}
	switch strings.ToLower(filepath.Ext(p.Name)) {
	case ".jpg", ".jpeg", ".png", ".pdf":
		return nil
	}
	return apperr.New(http.StatusBadRequest, "proof file must be jpg, png or pdf")
}

type Status string

const (
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
	StatusFailed  Status = "failed"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusPaid, StatusFailed:
		return true
	}
	return false
}

type Payment struct {
	ID           uuid.UUID       `db:"id" json:"id"`
	HouseholdID  uuid.UUID       `db:"household_id" json:"household_id"`
	WasteID      uuid.UUID       `db:"waste_id" json:"waste_id"`
	Amount       decimal.Decimal `db:"amount" json:"amount"`
	PaymentDate  *time.Time      `db:"payment_date" json:"payment_date"`
	Status       Status          `db:"status" json:"status"`
	ProofFileURL *string         `db:"proof_file_url" json:"proof_file_url"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at" json:"updated_at"`
}

type CreateRequest struct {
	HouseholdID uuid.UUID       `json:"household_id"`
	WasteID     uuid.UUID       `json:"waste_id"`
	Amount      decimal.Decimal `json:"amount"`
}

func (r CreateRequest) Validate() error {
	if r.HouseholdID == uuid.Nil {
		return apperr.New(http.StatusBadRequest, "household_id is required")
	}
	if r.WasteID == uuid.Nil {
		return apperr.New(http.StatusBadRequest, "waste_id is required")
	}
	if !r.Amount.IsPositive() {
		return apperr.New(http.StatusBadRequest, "amount must be greater than zero")
	}
	return nil
}

type ListFilter struct {
	Status      string
	HouseholdID uuid.UUID
	From        *time.Time
	To          *time.Time
	Page        httpres.Page
}

func ParseListFilter(r *http.Request) (ListFilter, error) {
	f := ListFilter{Page: httpres.ParsePage(r)}
	q := r.URL.Query()

	if s := q.Get("status"); s != "" {
		if !Status(s).Valid() {
			return f, apperr.New(http.StatusBadRequest, "invalid status filter")
		}
		f.Status = s
	}
	if h := q.Get("household_id"); h != "" {
		id, err := uuid.Parse(h)
		if err != nil {
			return f, apperr.New(http.StatusBadRequest, "invalid household_id filter")
		}
		f.HouseholdID = id
	}
	if v := q.Get("date_from"); v != "" {
		t, _, err := parseDate(v)
		if err != nil {
			return f, apperr.New(http.StatusBadRequest, "invalid date_from, use YYYY-MM-DD or RFC3339")
		}
		f.From = &t
	}
	if v := q.Get("date_to"); v != "" {
		t, dateOnly, err := parseDate(v)
		if err != nil {
			return f, apperr.New(http.StatusBadRequest, "invalid date_to, use YYYY-MM-DD or RFC3339")
		}
		if dateOnly {
			t = t.Add(24 * time.Hour)
		}
		f.To = &t
	}
	return f, nil
}

func parseDate(s string) (time.Time, bool, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, false, nil
	}
	t, err := time.Parse("2006-01-02", s)
	return t, true, err
}

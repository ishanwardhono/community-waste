package household

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ishanwardhono/community-waste/pkg/apperr"
)

type Household struct {
	ID        uuid.UUID `db:"id" json:"id"`
	OwnerName string    `db:"owner_name" json:"owner_name"`
	Address   string    `db:"address" json:"address"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreateRequest struct {
	OwnerName string `json:"owner_name"`
	Address   string `json:"address"`
}

func (r CreateRequest) Validate() error {
	if strings.TrimSpace(r.OwnerName) == "" {
		return apperr.New(http.StatusBadRequest, "owner_name is required")
	}
	if strings.TrimSpace(r.Address) == "" {
		return apperr.New(http.StatusBadRequest, "address is required")
	}
	return nil
}

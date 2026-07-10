package payment

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreateRequestValidate(t *testing.T) {
	hid, wid := uuid.New(), uuid.New()

	cases := []struct {
		name    string
		req     CreateRequest
		wantErr bool
	}{
		{"valid", CreateRequest{HouseholdID: hid, WasteID: wid, Amount: decimal.NewFromInt(50000)}, false},
		{"zero amount", CreateRequest{HouseholdID: hid, WasteID: wid}, true},
		{"negative amount", CreateRequest{HouseholdID: hid, WasteID: wid, Amount: decimal.NewFromInt(-1)}, true},
		{"missing waste", CreateRequest{HouseholdID: hid, Amount: decimal.NewFromInt(1)}, true},
		{"missing household", CreateRequest{WasteID: wid, Amount: decimal.NewFromInt(1)}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.req.Validate()
			if (err != nil) != c.wantErr {
				t.Fatalf("Validate() err = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}

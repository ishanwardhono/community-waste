package pickup

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateRequestValidate(t *testing.T) {
	hid := uuid.New()
	yes := true

	cases := []struct {
		name    string
		req     CreateRequest
		wantErr bool
	}{
		{"organic ok", CreateRequest{HouseholdID: hid, Type: TypeOrganic}, false},
		{"electronic with safety", CreateRequest{HouseholdID: hid, Type: TypeElectronic, SafetyCheck: &yes}, false},
		{"electronic without safety", CreateRequest{HouseholdID: hid, Type: TypeElectronic}, true},
		{"bad type", CreateRequest{HouseholdID: hid, Type: "glass"}, true},
		{"missing household", CreateRequest{Type: TypeOrganic}, true},
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

func TestScheduleRequestValidate(t *testing.T) {
	if err := (ScheduleRequest{}).Validate(); err == nil {
		t.Fatal("empty date should fail")
	}
	past := ScheduleRequest{PickupDate: time.Now().Add(-time.Hour)}
	if err := past.Validate(); err == nil {
		t.Fatal("past date should fail")
	}
	ok := ScheduleRequest{PickupDate: time.Now().Add(24 * time.Hour)}
	if err := ok.Validate(); err != nil {
		t.Fatalf("future date failed: %v", err)
	}
}

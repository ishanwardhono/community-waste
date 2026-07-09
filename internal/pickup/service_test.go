package pickup_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/ishanwardhono/community-waste/internal/household"
	"github.com/ishanwardhono/community-waste/internal/pickup"
	"github.com/ishanwardhono/community-waste/pkg/apperr"
	mockhousehold "github.com/ishanwardhono/community-waste/test/mocks/household"
	mockpickup "github.com/ishanwardhono/community-waste/test/mocks/pickup"
)

type pickupMocks struct {
	repo       *mockpickup.MockRepository
	households *mockhousehold.MockService
}

func newPickupService(t *testing.T) (pickup.Service, pickupMocks) {
	ctrl := gomock.NewController(t)
	m := pickupMocks{
		repo:       mockpickup.NewMockRepository(ctrl),
		households: mockhousehold.NewMockService(ctrl),
	}
	return pickup.NewService(m.repo, m.households), m
}

func TestCreateStartsPending(t *testing.T) {
	svc, m := newPickupService(t)
	req := validCreateRequest()

	m.households.EXPECT().Get(gomock.Any(), req.HouseholdID).Return(household.Household{}, nil)

	var saved pickup.Pickup
	m.repo.EXPECT().Insert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, p pickup.Pickup) error {
			saved = p
			return nil
		})

	got, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != pickup.StatusPending {
		t.Fatalf("status = %s, want pending", got.Status)
	}
	if saved.ID != got.ID {
		t.Fatal("returned pickup differs from saved pickup")
	}
}

func TestCreateUnknownHousehold(t *testing.T) {
	svc, m := newPickupService(t)
	req := validCreateRequest()

	m.households.EXPECT().Get(gomock.Any(), req.HouseholdID).
		Return(household.Household{}, apperr.New(http.StatusNotFound, "household not found"))

	_, err := svc.Create(context.Background(), req)
	assertCode(t, err, http.StatusNotFound)
}

func TestScheduleElectronicNeedsSafetyCheck(t *testing.T) {
	svc, m := newPickupService(t)
	id := uuidFor("22222222")
	no := false

	m.repo.EXPECT().Get(gomock.Any(), id).Return(pickup.Pickup{
		ID: id, Type: pickup.TypeElectronic, Status: pickup.StatusPending, SafetyCheck: &no,
	}, nil)

	_, err := svc.Schedule(context.Background(), id, pickup.ScheduleRequest{PickupDate: time.Now().Add(time.Hour)})
	assertCode(t, err, http.StatusUnprocessableEntity)
}

func TestScheduleHappyPath(t *testing.T) {
	svc, m := newPickupService(t)
	id := uuidFor("22222222")
	date := time.Now().Add(24 * time.Hour)

	m.repo.EXPECT().Get(gomock.Any(), id).Return(pickup.Pickup{
		ID: id, Type: pickup.TypeOrganic, Status: pickup.StatusPending,
	}, nil)
	m.repo.EXPECT().Schedule(gomock.Any(), id, date).Return(pickup.Pickup{
		ID: id, Status: pickup.StatusScheduled, PickupDate: &date,
	}, nil)

	got, err := svc.Schedule(context.Background(), id, pickup.ScheduleRequest{PickupDate: date})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != pickup.StatusScheduled {
		t.Fatalf("status = %s", got.Status)
	}
}

func TestCancelPassesThroughRepoConflict(t *testing.T) {
	svc, m := newPickupService(t)
	id := uuidFor("22222222")

	m.repo.EXPECT().Get(gomock.Any(), id).Return(pickup.Pickup{ID: id, Status: pickup.StatusCompleted}, nil)
	m.repo.EXPECT().Cancel(gomock.Any(), id).
		Return(pickup.Pickup{}, apperr.New(http.StatusConflict, "pickup can not be canceled from its current status"))

	_, err := svc.Cancel(context.Background(), id)
	assertCode(t, err, http.StatusConflict)
}

func validCreateRequest() pickup.CreateRequest {
	return pickup.CreateRequest{HouseholdID: uuidFor("11111111"), Type: pickup.TypeOrganic}
}

func assertCode(t *testing.T, err error, want int) {
	t.Helper()
	app, ok := err.(*apperr.AppError)
	if !ok {
		t.Fatalf("err = %v, want AppError", err)
	}
	if app.Code != want {
		t.Fatalf("code = %d, want %d", app.Code, want)
	}
}

func uuidFor(prefix string) uuid.UUID {
	return uuid.MustParse(prefix + "-0000-0000-0000-000000000000")
}

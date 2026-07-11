package payment_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"

	"github.com/ishanwardhono/community-waste/internal/household"
	"github.com/ishanwardhono/community-waste/internal/payment"
	"github.com/ishanwardhono/community-waste/internal/pickup"
	"github.com/ishanwardhono/community-waste/pkg/apperr"
	mockhousehold "github.com/ishanwardhono/community-waste/test/mocks/household"
	mockpayment "github.com/ishanwardhono/community-waste/test/mocks/payment"
	mockpickup "github.com/ishanwardhono/community-waste/test/mocks/pickup"
	mockstorage "github.com/ishanwardhono/community-waste/test/mocks/storage"
)

type paymentMocks struct {
	repo       *mockpayment.MockRepository
	households *mockhousehold.MockService
	pickups    *mockpickup.MockRepository
	store      *mockstorage.MockFileStorage
}

func newPaymentService(t *testing.T) (payment.Service, paymentMocks) {
	ctrl := gomock.NewController(t)
	m := paymentMocks{
		repo:       mockpayment.NewMockRepository(ctrl),
		households: mockhousehold.NewMockService(ctrl),
		pickups:    mockpickup.NewMockRepository(ctrl),
		store:      mockstorage.NewMockFileStorage(ctrl),
	}
	return payment.NewService(m.repo, m.households, m.pickups, m.store), m
}

func TestCreateRejectsForeignPickup(t *testing.T) {
	svc, m := newPaymentService(t)
	hid, otherHid, wid := uuid.New(), uuid.New(), uuid.New()

	m.households.EXPECT().Get(gomock.Any(), hid).Return(household.Household{ID: hid}, nil)
	m.pickups.EXPECT().Get(gomock.Any(), wid).Return(pickup.Pickup{ID: wid, HouseholdID: otherHid}, nil)

	_, err := svc.Create(context.Background(), payment.CreateRequest{
		HouseholdID: hid, WasteID: wid, Amount: decimal.NewFromInt(50000),
	})
	assertCode(t, err, http.StatusUnprocessableEntity)
}

func TestCreateUnknownHousehold(t *testing.T) {
	svc, m := newPaymentService(t)
	hid, wid := uuid.New(), uuid.New()

	m.households.EXPECT().Get(gomock.Any(), hid).
		Return(household.Household{}, apperr.New(http.StatusNotFound, "household not found"))

	_, err := svc.Create(context.Background(), payment.CreateRequest{
		HouseholdID: hid, WasteID: wid, Amount: decimal.NewFromInt(50000),
	})
	assertCode(t, err, http.StatusNotFound)
}

func TestCreateUnknownPickup(t *testing.T) {
	svc, m := newPaymentService(t)
	hid, wid := uuid.New(), uuid.New()

	m.households.EXPECT().Get(gomock.Any(), hid).Return(household.Household{ID: hid}, nil)
	m.pickups.EXPECT().Get(gomock.Any(), wid).
		Return(pickup.Pickup{}, apperr.New(http.StatusNotFound, "pickup not found"))

	_, err := svc.Create(context.Background(), payment.CreateRequest{
		HouseholdID: hid, WasteID: wid, Amount: decimal.NewFromInt(50000),
	})
	assertCode(t, err, http.StatusNotFound)
}

func TestCreateStartsPending(t *testing.T) {
	svc, m := newPaymentService(t)
	hid, wid := uuid.New(), uuid.New()

	m.households.EXPECT().Get(gomock.Any(), hid).Return(household.Household{ID: hid}, nil)
	m.pickups.EXPECT().Get(gomock.Any(), wid).Return(pickup.Pickup{ID: wid, HouseholdID: hid}, nil)

	var saved payment.Payment
	m.repo.EXPECT().Insert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, p payment.Payment) error {
			saved = p
			return nil
		})

	got, err := svc.Create(context.Background(), payment.CreateRequest{
		HouseholdID: hid, WasteID: wid, Amount: decimal.NewFromInt(50000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != payment.StatusPending {
		t.Fatalf("status = %s, want pending", got.Status)
	}
	if got.ID != saved.ID || got.ID == uuid.Nil {
		t.Fatalf("id not generated: %v", got.ID)
	}
	if !saved.Amount.Equal(decimal.NewFromInt(50000)) {
		t.Fatalf("amount = %s", saved.Amount)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatal("timestamps not set")
	}
}

func TestCreateForPickupAmounts(t *testing.T) {
	cases := []struct {
		wasteType pickup.WasteType
		want      int64
	}{
		{pickup.TypeOrganic, 50000},
		{pickup.TypePlastic, 50000},
		{pickup.TypePaper, 50000},
		{pickup.TypeElectronic, 100000},
	}
	for _, c := range cases {
		t.Run(string(c.wasteType), func(t *testing.T) {
			svc, m := newPaymentService(t)
			hid, wid := uuid.New(), uuid.New()

			var saved payment.Payment
			m.repo.EXPECT().Insert(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, p payment.Payment) error {
					saved = p
					return nil
				})

			if err := svc.CreateForPickup(context.Background(), hid, wid, c.wasteType); err != nil {
				t.Fatal(err)
			}
			if !saved.Amount.Equal(decimal.NewFromInt(c.want)) {
				t.Fatalf("amount = %s, want %d", saved.Amount, c.want)
			}
			if saved.Status != payment.StatusPending || saved.WasteID != wid || saved.HouseholdID != hid {
				t.Fatalf("bad payment: %+v", saved)
			}
		})
	}
}

func TestConfirmOnlyFromPending(t *testing.T) {
	svc, m := newPaymentService(t)
	id := uuid.New()

	m.repo.EXPECT().Get(gomock.Any(), id).Return(payment.Payment{ID: id, Status: payment.StatusPaid}, nil)

	_, err := svc.Confirm(context.Background(), id, payment.ProofFile{Name: "a.jpg", Size: 10})
	assertCode(t, err, http.StatusConflict)
}

func TestConfirmUploadsAndSavesURL(t *testing.T) {
	svc, m := newPaymentService(t)
	id := uuid.New()
	url := "http://localhost:9000/payment-proofs/proofs/x.jpg"

	m.repo.EXPECT().Get(gomock.Any(), id).Return(payment.Payment{ID: id, Status: payment.StatusPending}, nil)
	m.store.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), int64(10), "image/jpeg").Return(url, nil)
	m.repo.EXPECT().Confirm(gomock.Any(), id, url).Return(payment.Payment{
		ID: id, Status: payment.StatusPaid, ProofFileURL: &url,
	}, nil)

	got, err := svc.Confirm(context.Background(), id, payment.ProofFile{
		Name: "a.jpg", Size: 10, ContentType: "image/jpeg", Reader: strings.NewReader("x"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != payment.StatusPaid || got.ProofFileURL == nil {
		t.Fatalf("payment not confirmed: %+v", got)
	}
}

func TestConfirmStorageFailure(t *testing.T) {
	svc, m := newPaymentService(t)
	id := uuid.New()

	m.repo.EXPECT().Get(gomock.Any(), id).Return(payment.Payment{ID: id, Status: payment.StatusPending}, nil)
	m.store.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errors.New("minio down"))

	_, err := svc.Confirm(context.Background(), id, payment.ProofFile{Name: "a.jpg", Size: 10})
	if err == nil {
		t.Fatal("expected error")
	}
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

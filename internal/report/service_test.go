package report_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"

	"github.com/ishanwardhono/community-waste/internal/report"
	mockreport "github.com/ishanwardhono/community-waste/test/mocks/report"
)

func TestPaymentSummaryCountsOnlyPaidRevenue(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mockreport.NewMockRepository(ctrl)
	svc := report.NewService(repo)

	repo.EXPECT().PaymentSummary(gomock.Any()).Return([]report.PaymentStatusRow{
		{Status: "paid", Count: 3, TotalAmount: decimal.NewFromInt(250000)},
		{Status: "pending", Count: 2, TotalAmount: decimal.NewFromInt(100000)},
		{Status: "failed", Count: 1, TotalAmount: decimal.NewFromInt(50000)},
	}, nil)

	got, err := svc.PaymentSummary(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.TotalRevenue.Equal(decimal.NewFromInt(250000)) {
		t.Fatalf("revenue = %s, want 250000", got.TotalRevenue)
	}
	if len(got.ByStatus) != 3 {
		t.Fatalf("rows = %d", len(got.ByStatus))
	}
}

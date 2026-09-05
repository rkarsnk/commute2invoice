package service_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/rkarsnk/commute2invoice/src/db"
	"github.com/rkarsnk/commute2invoice/src/models"
	"github.com/rkarsnk/commute2invoice/src/repository"
	"github.com/rkarsnk/commute2invoice/src/service"
)

type testServices struct {
	fares    *service.FareService
	trips    *service.TripService
	invoices *service.InvoiceService
	repo     repository.Repository
}

func newTestServices(t *testing.T) testServices {
	t.Helper()

	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("initialize database: %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(database); err != nil {
			t.Errorf("close database: %v", err)
		}
	})

	repo := repository.NewSQLiteRepository(database)
	return testServices{
		fares:    service.NewFareService(repo),
		trips:    service.NewTripService(repo, repo),
		invoices: service.NewInvoiceService(repo, repo, repo),
		repo:     repo,
	}
}

func createFare(t *testing.T, fares *service.FareService, validFrom string, validUntil *string, oneWay int, roundTrip *int, image []byte) int {
	t.Helper()

	id, err := fares.CreateFareEvidence(context.Background(), &models.CreateFareEvidenceRequest{
		ImageID:       "test-image",
		ImageMime:     "image/png",
		ImageData:     base64.StdEncoding.EncodeToString(image),
		Label:         "Test route",
		CompanyText:   "Test Railway",
		FromText:      "Origin",
		ToText:        "Destination",
		FareOneway:    oneWay,
		FareRoundtrip: roundTrip,
		ValidFrom:     validFrom,
		ValidUntil:    validUntil,
	})
	if err != nil {
		t.Fatalf("create fare evidence: %v", err)
	}
	return id
}

func createTrip(t *testing.T, trips *service.TripService, date string, fareID int, roundTrip bool, billingTarget string) int {
	t.Helper()

	id, err := trips.CreateTrip(context.Background(), date, &models.CreateTripRequest{
		FareEvidenceID: fareID,
		RoundTrip:      roundTrip,
		BillingTarget:  billingTarget,
	})
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}
	return id
}

func TestUpdateFareEvidenceValidityRejectsDateBeforeStartAndPersistsBoundary(t *testing.T) {
	s := newTestServices(t)
	ctx := context.Background()
	fareID := createFare(t, s.fares, "2024-04-01", nil, 100, nil, []byte("fare-image"))

	invalidUntil := "2024-03-31"
	err := s.fares.UpdateFareEvidenceValidity(ctx, fareID, &invalidUntil)
	if err == nil || !strings.Contains(err.Error(), "must not be before valid_from") {
		t.Fatalf("expected invalid validity range error, got %v", err)
	}

	fare, err := s.fares.GetFareEvidenceByID(ctx, fareID)
	if err != nil {
		t.Fatalf("get fare evidence after rejected update: %v", err)
	}
	if fare.ValidUntil != nil {
		t.Fatalf("rejected update changed valid_until to %q", *fare.ValidUntil)
	}

	boundary := "2024-04-30"
	if err := s.fares.UpdateFareEvidenceValidity(ctx, fareID, &boundary); err != nil {
		t.Fatalf("set valid_until at valid range boundary: %v", err)
	}
	valid, err := s.fares.ListValidFareEvidenceForDate(ctx, "2024-04-30")
	if err != nil {
		t.Fatalf("list fares on validity boundary: %v", err)
	}
	if len(valid) != 1 || valid[0].ID != fareID {
		t.Fatalf("expected fare to be valid through boundary, got %#v", valid)
	}
	valid, err = s.fares.ListValidFareEvidenceForDate(ctx, "2024-05-01")
	if err != nil {
		t.Fatalf("list fares after validity boundary: %v", err)
	}
	if len(valid) != 0 {
		t.Fatalf("expected no fares after validity boundary, got %#v", valid)
	}
}

func TestDeleteFareEvidenceRejectsReferencedTrip(t *testing.T) {
	s := newTestServices(t)
	ctx := context.Background()
	fareID := createFare(t, s.fares, "2024-01-01", nil, 100, nil, []byte("fare-image"))
	tripID := createTrip(t, s.trips, "2024-04-01", fareID, false, "自社")

	err := s.fares.DeleteFareEvidence(ctx, fareID)
	if err == nil || !strings.Contains(err.Error(), "used by 1 trip") {
		t.Fatalf("expected referenced fare deletion to fail, got %v", err)
	}
	if _, err := s.fares.GetFareEvidenceByID(ctx, fareID); err != nil {
		t.Fatalf("referenced fare was deleted: %v", err)
	}

	if err := s.trips.DeleteTrip(ctx, tripID); err != nil {
		t.Fatalf("delete referencing trip: %v", err)
	}
	if err := s.fares.DeleteFareEvidence(ctx, fareID); err != nil {
		t.Fatalf("delete unreferenced fare: %v", err)
	}
}

func TestGetCalendarDaysAggregatesEveryDayInLeapMonth(t *testing.T) {
	s := newTestServices(t)
	ctx := context.Background()
	fareID := createFare(t, s.fares, "2024-01-01", nil, 100, nil, []byte("fare-image"))
	createTrip(t, s.trips, "2024-02-01", fareID, false, "自社")
	createTrip(t, s.trips, "2024-02-01", fareID, true, "客先")
	createTrip(t, s.trips, "2024-02-29", fareID, false, "自社")
	createTrip(t, s.trips, "2024-03-01", fareID, false, "自社")

	days, err := s.trips.GetCalendarDays(ctx, 2024, 2)
	if err != nil {
		t.Fatalf("get calendar days: %v", err)
	}
	if len(days) != 29 {
		t.Fatalf("expected 29 days in February 2024, got %d", len(days))
	}
	if days[0].Date != "2024-02-01" || days[0].TripCount != 2 {
		t.Fatalf("unexpected first day aggregation: %#v", days[0])
	}
	if days[28].Date != "2024-02-29" || days[28].TripCount != 1 {
		t.Fatalf("unexpected leap day aggregation: %#v", days[28])
	}
}

func TestUpdateTripChangesOnlyEditableFields(t *testing.T) {
	s := newTestServices(t)
	ctx := context.Background()
	fareID := createFare(t, s.fares, "2024-01-01", nil, 100, nil, []byte("fare-image"))
	tripID := createTrip(t, s.trips, "2024-02-29", fareID, false, "自社")

	if err := s.trips.UpdateTrip(ctx, tripID, &models.UpdateTripRequest{
		RoundTrip:     true,
		BillingTarget: "客先",
	}); err != nil {
		t.Fatalf("update trip: %v", err)
	}

	trip, err := s.trips.GetTripByID(ctx, tripID)
	if err != nil {
		t.Fatalf("get updated trip: %v", err)
	}
	if !trip.RoundTrip || trip.BillingTarget != "客先" {
		t.Fatalf("trip fields were not updated: %#v", trip)
	}
	if trip.Date != "2024-02-29" || trip.FareEvidenceID != fareID {
		t.Fatalf("immutable trip fields changed: %#v", trip)
	}

	err = s.trips.UpdateTrip(ctx, tripID, &models.UpdateTripRequest{BillingTarget: "invalid"})
	if err == nil || !strings.Contains(err.Error(), "invalid billing_target") {
		t.Fatalf("expected invalid billing target error, got %v", err)
	}
}

func TestInvoicePreviewAndStoredPDFSnapshot(t *testing.T) {
	s := newTestServices(t)
	ctx := context.Background()
	roundTripFare := 175
	firstFareID := createFare(t, s.fares, "2024-01-01", nil, 100, &roundTripFare, []byte("first-image"))
	secondFareID := createFare(t, s.fares, "2024-01-01", nil, 200, nil, []byte("second-image"))
	createTrip(t, s.trips, "2024-04-02", firstFareID, true, "自社")
	createTrip(t, s.trips, "2024-04-03", secondFareID, false, "自社")
	createTrip(t, s.trips, "2024-04-04", firstFareID, false, "客先")
	createTrip(t, s.trips, "2024-05-01", firstFareID, false, "自社")

	preview, err := s.invoices.GetInvoicePreview(ctx, 2024, 4, "自社")
	if err != nil {
		t.Fatalf("get invoice preview: %v", err)
	}
	if preview.Status != "draft" || len(preview.InvoiceLines) != 2 || preview.TotalFare != 375 {
		t.Fatalf("unexpected invoice preview: %#v", preview)
	}
	if len(preview.EvidenceImages) != 2 {
		t.Fatalf("expected images for two used fares, got %#v", preview.EvidenceImages)
	}
	images := make(map[int]string, len(preview.EvidenceImages))
	for _, image := range preview.EvidenceImages {
		images[image.FareEvidenceID] = image.ImageBase64
	}
	if images[firstFareID] != base64.StdEncoding.EncodeToString([]byte("first-image")) ||
		images[secondFareID] != base64.StdEncoding.EncodeToString([]byte("second-image")) {
		t.Fatalf("preview evidence images do not match stored images: %#v", images)
	}

	pdfSnapshot := []byte("%PDF-test-snapshot")
	invoiceID, err := s.repo.CreateInvoice(ctx, &models.Invoice{
		YearMonth:     "2024-04",
		BillingTarget: "自社",
		PDF:           pdfSnapshot,
	})
	if err != nil {
		t.Fatalf("store invoice PDF snapshot: %v", err)
	}
	stored, err := s.repo.GetInvoiceByYearMonthAndBillingTarget(ctx, 2024, 4, "自社")
	if err != nil {
		t.Fatalf("get stored invoice: %v", err)
	}
	if stored.ID != invoiceID || !bytes.Equal(stored.PDF, pdfSnapshot) {
		t.Fatalf("stored invoice differs from snapshot: %#v", stored)
	}

	confirmedPreview, err := s.invoices.GetInvoicePreview(ctx, 2024, 4, "自社")
	if err != nil {
		t.Fatalf("get confirmed invoice preview: %v", err)
	}
	if confirmedPreview.Status != "confirmed" || confirmedPreview.ConfirmedAt == nil ||
		*confirmedPreview.ConfirmedAt != stored.GeneratedAt {
		t.Fatalf("invoice confirmation state is wrong: %#v", confirmedPreview)
	}
	pdf, err := s.invoices.GetInvoicePDF(ctx, 2024, 4, "自社")
	if err != nil {
		t.Fatalf("get invoice PDF: %v", err)
	}
	if !bytes.Equal(pdf, pdfSnapshot) {
		t.Fatalf("returned PDF differs from stored snapshot: %q", pdf)
	}
}

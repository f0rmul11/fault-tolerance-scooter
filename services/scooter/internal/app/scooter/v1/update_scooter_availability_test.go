package v1

import (
	"testing"

	pb "scooter/internal/pkg/pb/scooter/scooter/v1"
	app "scooter/internal/pkg/scootererror"
)

func TestAvailabilityRequiresExplicitFlag(t *testing.T) {
	_, err := New(nil).UpdateScooterAvailability(t.Context(), &pb.UpdateScooterAvailabilityRequest{})
	if !app.IsCode(err, app.InvalidArgument) {
		t.Fatalf("absent flag must not deactivate scooter: %v", err)
	}
}

package loyalty

// DiscountPercent определяет скидку по числу завершённых поездок.
func DiscountPercent(completedTrips int64) int32 {
	switch {
	case completedTrips >= 50:
		return 20
	case completedTrips >= 10:
		return 10
	default:
		return 0
	}
}

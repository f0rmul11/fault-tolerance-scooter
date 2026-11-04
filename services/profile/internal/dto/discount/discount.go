package discount

// Discount описывает скидку для отображения в личном кабинете.
type Discount struct {
	Percent             int32
	CompletedTripsCount *int64
	Degraded            bool
	Source              string
}

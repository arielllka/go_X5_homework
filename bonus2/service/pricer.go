package service

type DeliveryPricer interface {
	Price(weightGrams, distanceKM int) (int64, error)
}

type BaseTariff struct {
	BasePriceKopecks int64
	PerKMKopecks     int64
	MaxDistanceKM    int
}

func (b BaseTariff) validate(weightGrams, distanceKM int) error {
	if weightGrams <= 0 {
		return ErrInvalidWeight
	}
	if distanceKM <= 0 {
		return ErrInvalidDistance
	}
	if distanceKM > b.MaxDistanceKM {
		return ErrDistanceTooLarge
	}
	return nil
}

func billableKilograms(weightGrams int) int64 {
	kilograms := int64(weightGrams / 1000)
	if weightGrams%1000 > 0 {
		kilograms++
	}
	return kilograms
}

type Courier struct {
	BaseTariff
	PerKGKopecks int64
}

func NewCourier() Courier {
	return Courier{BaseTariff: BaseTariff{20_000, 1_500, 100}, PerKGKopecks: 5_000}
}

func (t Courier) Price(weightGrams, distanceKM int) (int64, error) {
	if err := t.validate(weightGrams, distanceKM); err != nil {
		return 0, err
	}
	return t.BasePriceKopecks + int64(distanceKM)*t.PerKMKopecks + billableKilograms(weightGrams)*t.PerKGKopecks, nil
}

type Pickup struct {
	BaseTariff
	PerKGKopecks int64
}

func NewPickup() Pickup {
	return Pickup{BaseTariff: BaseTariff{10_000, 500, 50}, PerKGKopecks: 2_000}
}

func (t Pickup) Price(weightGrams, distanceKM int) (int64, error) {
	if err := t.validate(weightGrams, distanceKM); err != nil {
		return 0, err
	}
	return t.BasePriceKopecks + int64(distanceKM)*t.PerKMKopecks + billableKilograms(weightGrams)*t.PerKGKopecks, nil
}

type Express struct {
	BaseTariff
	PerKGKopecks int64
}

func NewExpress() Express {
	return Express{BaseTariff: BaseTariff{50_000, 3_000, 300}, PerKGKopecks: 10_000}
}

func (t Express) Price(weightGrams, distanceKM int) (int64, error) {
	if err := t.validate(weightGrams, distanceKM); err != nil {
		return 0, err
	}
	return t.BasePriceKopecks + int64(distanceKM)*t.PerKMKopecks + billableKilograms(weightGrams)*t.PerKGKopecks, nil
}

var (
	_ DeliveryPricer = Courier{}
	_ DeliveryPricer = Pickup{}
	_ DeliveryPricer = Express{}
)

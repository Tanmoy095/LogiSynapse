package pricing

import "errors"

type Tier struct {
	UpTo     int   // Maximum distance (in miles) for this tier...inclusive upper limit
	UnitCost int64 //cost in center per unit
	FlatFee  int64 // flat fee in cents for this tier
}

type PricingEngine struct {
	Tiers []Tier // Ordered list of pricing tiers
}

func (e *PricingEngine) ValidateTiers() error {
	if len(e.Tiers) == 0 {
		return errors.New("at least one tier is required")
	}

	// Last tier must be unlimited (-1)
	if e.Tiers[len(e.Tiers)-1].UpTo != -1 {
		return errors.New("last tier must be unlimited (-1)")
	}

	// Validate pairwise ordering and that no tiers follow an unlimited tier
	for i := 1; i < len(e.Tiers); i++ {
		prev := e.Tiers[i-1]
		current := e.Tiers[i]

		if prev.UpTo == -1 {
			return errors.New("no tiers allowed after unlimited tier")
		}

		if current.UpTo != -1 && current.UpTo <= prev.UpTo {
			return errors.New("tiers must be strictly increasing")
		}
	}

	// Ensure costs are non-negative
	for _, tier := range e.Tiers {
		if tier.UnitCost < 0 || tier.FlatFee < 0 {
			return errors.New("costs must be non-negative")
		}
	}

	return nil
}

func (e *PricingEngine) CalculateCost(usage int) (int64, error) {
	if usage < 0 {
		return 0, errors.New("usage cannot be negative")
	}

	var totalCost int64 = 0
	remainingUsage := usage
	prevLimit := 0

	for _, tier := range e.Tiers {
		if remainingUsage <= 0 {
			break // No more usage to allocate
		}

		// Determine how many units fall into this tier (explicit inclusive)
		var tierUsage int
		if tier.UpTo == -1 {
			// Unlimited last tier: take all remaining
			tierUsage = remainingUsage
		} else {
			// Capacity: from prevLimit+1 to UpTo inclusive
			capacity := tier.UpTo - prevLimit
			if remainingUsage > capacity {
				tierUsage = capacity
			} else {
				tierUsage = remainingUsage
			}
		}

		// Add per-unit cost
		totalCost += int64(tierUsage) * tier.UnitCost
		// Add FlatFee only if tier entered (usage >0)
		if tierUsage > 0 {
			totalCost += tier.FlatFee
		}

		// Update tracking
		remainingUsage -= tierUsage
		if tier.UpTo != -1 {
			prevLimit = tier.UpTo // Next starts at UpTo+1
		}
	}

	// If remaining >0, validation failed (no unlimited)—but shouldn't happen
	if remainingUsage > 0 {
		return 0, errors.New("usage exceeds tiers (validation should prevent)")
	}
	return totalCost, nil
}

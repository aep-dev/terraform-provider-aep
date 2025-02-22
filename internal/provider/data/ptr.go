package data

import "math/big"

// Helper functions for creating pointers.
func String(v string) *string {
	return &v
}

func BigFloat(v *big.Float) *big.Float {
	return v
}

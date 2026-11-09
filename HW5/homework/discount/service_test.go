package discount

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent int
		wantErr error
	}{
		{"zero", 0, nil},
		{"ten", 10, nil},
		{"hundred", 100, nil},
		{"negative", -1, ErrInvalidPercent},
		{"above_hundred", 101, ErrInvalidPercent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePercent(tt.percent)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestApply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		price   int
		percent int
		want    int
		wantErr error
	}{
		{"regular_discount", 1000, 15, 850, nil},
		{"zero_discount", 1000, 0, 1000, nil},
		{"full_discount", 1000, 100, 0, nil},
		{"zero_price", 0, 25, 0, nil},
		{"integer_rounding", 3, 50, 1, nil},
		{"negative_price", -1, 10, 0, ErrNegativePrice},
		{"negative_percent", 1000, -1, 0, ErrInvalidPercent},
		{"percent_above_hundred", 1000, 101, 0, ErrInvalidPercent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Apply(tt.price, tt.percent)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFinalPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		items   []Item
		want    int
		wantErr error
	}{
		{"nil_list", nil, 0, nil},
		{"empty_list", []Item{}, 0, nil},
		{"several_items", []Item{
			{Price: 1000, DiscountPercent: 10},
			{Price: 2000, DiscountPercent: 25},
			{Price: 300, DiscountPercent: 100},
		}, 2400, nil},
		{"round_each_item", []Item{
			{Price: 3, DiscountPercent: 50},
			{Price: 3, DiscountPercent: 50},
		}, 2, nil},
		{"negative_price", []Item{
			{Price: 1000, DiscountPercent: 10},
			{Price: -1, DiscountPercent: 10},
		}, 0, ErrNegativePrice},
		{"negative_percent", []Item{
			{Price: 1000, DiscountPercent: 10},
			{Price: 1000, DiscountPercent: -1},
		}, 0, ErrInvalidPercent},
		{"percent_above_hundred", []Item{
			{Price: 1000, DiscountPercent: 10},
			{Price: 1000, DiscountPercent: 101},
		}, 0, ErrInvalidPercent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FinalPrice(tt.items)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

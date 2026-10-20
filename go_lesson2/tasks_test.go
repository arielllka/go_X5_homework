package lesson2

import (
	"fmt"
	"math"
	"testing"
)

func TestFizzBuzz(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{-1, "invalid"}, {0, "invalid"}, {1, ""}, {2, ""}, {3, "Fizz"}, {5, "Buzz"},
		{6, "Fizz"}, {10, "Buzz"}, {15, "FizzBuzz"}, {30, "FizzBuzz"}, {31, ""},
	}
	for _, c := range cases {
		t.Run(fmt.Sprint(c.n), func(t *testing.T) {
			if got := FizzBuzz(c.n); got != c.want {
				t.Errorf("FizzBuzz(%d) = %q, ожидается %q", c.n, got, c.want)
			}
		})
	}
}

func TestStatistics(t *testing.T) {
	cases := []struct {
		name          string
		values        []int64
		min, max, sum int64
		avg           float64
	}{
		{"nil", nil, 0, 0, 0, 0},
		{"empty", []int64{}, 0, 0, 0, 0},
		{"orders", []int64{12000, 7500, 31000, 4500}, 4500, 31000, 55000, 13750},
		{"one", []int64{42}, 42, 42, 42, 42},
		{"negative", []int64{-8, -2, -5}, -8, -2, -15, -5},
		{"mixed", []int64{-5, 0, 9}, -5, 9, 4, 4.0 / 3.0},
		{"fraction", []int64{1, 2}, 1, 2, 3, 1.5},
		{"zeros", []int64{0, 0}, 0, 0, 0, 0},
		{"equal", []int64{7, 7, 7}, 7, 7, 21, 7},
		{"large", []int64{3000000000, 4000000000}, 3000000000, 4000000000, 7000000000, 3500000000},
	}

	for _, c := range cases {
		t.Run(c.name+"/Sum", func(t *testing.T) {
			if got := Sum(c.values); got != c.sum {
				t.Errorf("Sum(%v) = %d, ожидается %d", c.values, got, c.sum)
			}
		})
		t.Run(c.name+"/Min", func(t *testing.T) {
			if got := Min(c.values); got != c.min {
				t.Errorf("Min(%v) = %d, ожидается %d", c.values, got, c.min)
			}
		})
		t.Run(c.name+"/Max", func(t *testing.T) {
			if got := Max(c.values); got != c.max {
				t.Errorf("Max(%v) = %d, ожидается %d", c.values, got, c.max)
			}
		})
		t.Run(c.name+"/Average", func(t *testing.T) {
			got := Average(c.values)
			if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-c.avg) > 1e-9 {
				t.Errorf("Average(%v) = %v, ожидается %v", c.values, got, c.avg)
			}
		})
	}
}

func TestCategory(t *testing.T) {
	cases := []struct {
		amount int64
		want   string
	}{
		{-1, "invalid"}, {0, "small"}, {9999, "small"}, {10000, "regular"},
		{49999, "regular"}, {50000, "large"}, {50001, "large"}, {4000000000, "large"},
	}
	for _, c := range cases {
		t.Run(fmt.Sprint(c.amount), func(t *testing.T) {
			if got := Category(c.amount); got != c.want {
				t.Errorf("Category(%d) = %q, ожидается %q", c.amount, got, c.want)
			}
		})
	}
}

func TestRuneCount(t *testing.T) {
	cases := []struct {
		name, text string
		want       int
	}{
		{"empty", "", 0}, {"ascii", "Go", 2}, {"cyrillic", "Привет", 6},
		{"mixed", "Goшник", 6}, {"emoji", "🙂", 1}, {"combining", "e\u0301", 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := RuneCount(c.text); got != c.want {
				t.Errorf("RuneCount(%q) = %d, ожидается %d", c.text, got, c.want)
			}
		})
	}
}

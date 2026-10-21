package homework

import "testing"

func TestLongestPrefix(t *testing.T) {
	cases := []struct {
		name  string
		words []string
		want  string
	}{
		{"contest_basic_1", []string{"flower", "flow", "flight"}, "fl"},
		{"contest_basic_2", []string{"interview", "internet", "internal"}, "inter"},
		{"contest_single", []string{"a"}, "a"},
		{"contest_empty_word", []string{"", "backend"}, ""},
		{"contest_no_prefix", []string{"dog", "racecar", "car"}, ""},
		{"contest_set", []string{"prefix", "preform", "prevent", "preview", "pre"}, "pre"},
		{"extra_nil", nil, ""},
		{"extra_empty", []string{}, ""},
		{"extra_unicode", []string{"ёжик", "ёлка"}, "ё"},
		{"extra_different_runes", []string{"я", "ю"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := longestPrefix(c.words); got != c.want {
				t.Errorf("longestPrefix(%q) = %q, ожидается %q", c.words, got, c.want)
			}
		})
	}
}

func TestPairIndices(t *testing.T) {
	cases := []struct {
		name  string
		input PairInput
		want  [2]int
	}{
		{"contest_basic_1", PairInput{Nums: []int{2, 7, 11, 15}, Target: 9}, [2]int{0, 1}},
		{"contest_basic_2", PairInput{Nums: []int{3, 2, 4}, Target: 6}, [2]int{1, 2}},
		{"contest_equal_values", PairInput{Nums: []int{3, 3}, Target: 6}, [2]int{0, 1}},
		{"contest_negative", PairInput{Nums: []int{-4, -1, 2, 8}, Target: 7}, [2]int{1, 3}},
		{"contest_zeros", PairInput{Nums: []int{0, 4, 3, 0}, Target: 0}, [2]int{0, 3}},
		{"contest_set_corrected", PairInput{Nums: []int{100, 201, 306, 407, 509, 611, 713, 815}, Target: 916}, [2]int{3, 4}},
		{"extra_first_last", PairInput{Nums: []int{8, 1, 4, 2}, Target: 10}, [2]int{0, 3}},
		{"extra_negative_sum", PairInput{Nums: []int{-8, -5, -2}, Target: -10}, [2]int{0, 2}},
		{"extra_zero_positive", PairInput{Nums: []int{0, 3, 5}, Target: 3}, [2]int{0, 1}},
		{"extra_no_reuse", PairInput{Nums: []int{5, 1, 9}, Target: 10}, [2]int{1, 2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := pairIndices(c.input); got != c.want {
				t.Errorf("pairIndices(%+v) = %v, ожидается %v", c.input, got, c.want)
			}
		})
	}
}

func TestHasDuplicate(t *testing.T) {
	cases := []struct {
		name string
		nums []int
		want bool
	}{
		{"contest_basic_1", []int{1, 2, 3, 1}, true},
		{"contest_basic_2", []int{1, 2, 3, 4}, false},
		{"contest_single", []int{7}, false},
		{"contest_negative", []int{-1, -2, -1}, true},
		{"contest_repeat_end", []int{5, 4, 3, 2, 1, 5}, true},
		{"contest_set", []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, false},
		{"extra_nil", nil, false},
		{"extra_empty", []int{}, false},
		{"extra_three_equal", []int{4, 4, 4}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasDuplicate(c.nums); got != c.want {
				t.Errorf("hasDuplicate(%v) = %t, ожидается %t", c.nums, got, c.want)
			}
		})
	}
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCompactSorted(t *testing.T) {
	cases := []struct {
		name       string
		nums, want []int
	}{
		{"contest_basic_1", []int{1, 1, 2}, []int{1, 2}},
		{"contest_basic_2", []int{0, 0, 1, 1, 2, 2, 3}, []int{0, 1, 2, 3}},
		{"contest_empty", []int{}, []int{}},
		{"contest_unique", []int{-2, -1, 0, 1, 2}, []int{-2, -1, 0, 1, 2}},
		{"contest_all_equal", []int{5, 5, 5, 5, 5}, []int{5}},
		{"contest_set", []int{-3, -3, -2, -2, -1, 0, 0, 1, 1, 2, 3, 3}, []int{-3, -2, -1, 0, 1, 2, 3}},
		{"extra_nil", nil, []int{}},
		{"extra_single", []int{9}, []int{9}},
		{"extra_runs", []int{-5, -5, -5, 0, 0, 9}, []int{-5, 0, 9}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := append([]int(nil), c.nums...)
			got := compactSorted(c.nums)
			if !sameInts(got, c.want) {
				t.Fatalf("compactSorted(%v) = %v, ожидается %v", before, got, c.want)
			}
			if !sameInts(c.nums, before) {
				t.Fatal("изменён входной слайс")
			}
			if len(got) > 0 {
				saved := got[0]
				got[0] ^= 1
				if !sameInts(c.nums, before) {
					t.Error("результат использует общие элементы с входом, нужен новый слайс")
				}
				got[0] = saved
			}
		})
	}
}

func TestBalancedBrackets(t *testing.T) {
	cases := []struct {
		name, text string
		want       bool
	}{
		{"contest_basic_1", "()[]{}", true},
		{"contest_basic_2", "([]){}", true},
		{"contest_wrong_pair", "(]", false},
		{"contest_wrong_order", "([)]", false},
		{"contest_unclosed", "(([]){})(", false},
		{"contest_set", "{[()()]}[{}]", true},
		{"extra_empty", "", true},
		{"extra_open_only", "(((", false},
		{"extra_close_first", "}{", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := balancedBrackets(c.text); got != c.want {
				t.Errorf("balancedBrackets(%q) = %t, ожидается %t", c.text, got, c.want)
			}
		})
	}
}

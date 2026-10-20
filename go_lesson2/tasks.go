package lesson2

// FizzBuzz возвращает FizzBuzz, Fizz, Buzz или пустую строку для обычного числа.
// Для n <= 0 возвращает invalid.
func FizzBuzz(n int) string {
	switch {
	case n <= 0:
		return "invalid"

	case n%15 == 0:
		return "FizzBuzz"

	case n%5 == 0:
		return "Buzz"

	case n%3 == 0:
		return "Fizz"
	}
	return ""
}

// Sum возвращает сумму. Для пустого среза и nil возвращает 0.
func Sum(values []int64) int64 {
	if values == nil {
		return 0
	}

	if len(values) == 0 {
		return 0
	}
	var sum int64
	for _, value := range values {
		sum += value
	}

	return sum

}

// Min возвращает минимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Min(values []int64) int64 {

	if values == nil {
		return 0
	}

	if len(values) == 0 {
		return 0
	}

	min := values[0]

	for _, value := range values {
		if value < min {
			min = value
		}
	}

	return min

}

// Max возвращает максимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Max(values []int64) int64 {

	if values == nil {
		return 0
	}

	if len(values) == 0 {
		return 0
	}

	max := values[0]

	for _, value := range values {
		if value > max {
			max = value
		}
	}

	return max
}

// Average возвращает среднее как float64. Для пустого среза и nil возвращает 0.
func Average(values []int64) float64 {

	if values == nil {
		return 0
	}
	if len(values) == 0 {
		return 0
	}

	var sum float64 = 0
	var count float64

	for _, value := range values {
		sum += float64(value)
		count++
	}

	sum /= count

	return sum
}

// Category возвращает invalid, small, regular или large по сумме в копейках.
func Category(amount int64) string {
	switch {
	case amount < 0:
		return "invalid"
	case amount <= 9999:
		return "small"
	case amount <= 49999:
		return "regular"

	}
	return "large"
}

// RuneCount считает кодовые точки Unicode в строке через range.
func RuneCount(text string) int {
	count := 0

	for range text {
		count++
	}

	return count
}

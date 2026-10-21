package homework

// PairInput соответствует объекту nums/target из условия контеста.
type PairInput struct {
	Nums   []int
	Target int
}

// longestPrefix возвращает самый длинный общий префикс строк.
func longestPrefix(words []string) string {

	if (len(words) == 0){
		return ""
	}

	first := []rune(words[0])
	result := ""
	
	for j, letter := range first{
		for i := 0; i < len(words); i++ {

            current := []rune(words[i])
			if( (j >= len(current)) || current[j] != letter ){
				return result
			} 
		}
		result += string(letter)

	}
	return result
}

// pairIndices возвращает индексы двух разных элементов по возрастанию.
func pairIndices(input PairInput) [2]int {

	len_nums:= len(input.Nums)

	for i := 0; i < len_nums-1; i++ {

       for j := i + 1; j < len_nums; j++ {

		sum:= input.Nums[i] + input.Nums[j] 

		if( sum == input.Target){
			return [2]int{i, j}
		}

	   }
	}
	return [2]int{}
} 

// hasDuplicate сообщает, встречается ли какое-либо значение повторно.
func hasDuplicate(nums []int) bool {
    seen := make(map[int]bool)

    for _, num := range nums {
        if seen[num] {
            return true
        }

        seen[num] = true
    }

    return false
}

// compactSorted создаёт независимый слайс уникальных значений.
func compactSorted(nums []int) []int {
	// TODO: сохранить порядок и не изменять исходный слайс.
	if(len(nums) == 0){
		return nil
	}
	current := nums[0]
	result:= []int{}

	result = append(result, current)

	for _ , number := range nums[1:]{
		if (number != current){
			current = number
			result = append(result, number)

		}
	} 
	return result
}

// balancedBrackets проверяет пары и порядок закрытия скобок.
func balancedBrackets(text string) bool {
	stack := []rune{}
	cur_len:=0
	for _ , symbol := range text{ 
	switch symbol{

	case '(' , '{' , '[':
		stack = append(stack , symbol)
	
	case ')':
		cur_len = len(stack)
		if(cur_len == 0 || stack[cur_len-1] != '('){
			return false
		}
		stack = stack[:cur_len-1]

	case ']':
		cur_len = len(stack)
		if(cur_len == 0 || stack[cur_len-1] != '['){
			return false
		}
		stack = stack[:cur_len-1]
		
	case '}':
		cur_len = len(stack)
		if(cur_len == 0 || stack[cur_len-1] != '{'){
			return false
		}
		stack = stack[:cur_len-1]
	

	default: continue
	}

 }

return len(stack) == 0

}

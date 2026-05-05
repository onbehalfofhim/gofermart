package utils

import (
	"strconv"
	"strings"
)

// проверка номера заказа по алгоритму Луна
func ValidateLuhn(number string) bool {
	// Убираем пробелы, если они есть
	number = strings.ReplaceAll(number, " ", "")

	var sum int
	nDigits := len(number)
	parity := nDigits % 2

	for i, r := range number {
		digit, err := strconv.Atoi(string(r))
		if err != nil {
			return false // Некорректный символ
		}

		// Удваиваем каждую вторую цифру
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	// Сумма должна делиться на 10
	return sum%10 == 0
}

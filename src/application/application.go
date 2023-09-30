package application

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

var acceptedFormats = []string{
	time.RFC3339,
	time.DateOnly, // "2006-01-02"
	"2006/01/02",
	"02/01/2006",
	"02-01-2006",
}

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var err error
	s := string(b)
	s = s[1 : len(s)-1] // remove quotes
	*t, err = NewTime(s)
	return err
}

func (t Time) ToString() string {
	return t.Format(time.DateOnly)
}

var ErrDate = errors.New("Invalid date format")

func NewTime(dateString string) (Time, error) {
	var convertedTime time.Time
	var err error

	valid := false
	for _, acceptedLayout := range acceptedFormats {
		convertedTime, err = time.Parse(acceptedLayout, dateString)
		if err == nil {
			valid = true
			break
		}
	}

	if valid {
		return Time{convertedTime}, nil
	}
	return Time{convertedTime}, ErrDate
}

var ErrDescription = errors.New("Description exceeds 50 characters")

func NewDescription(description string) (string, error) {
	if len(description) > 50 {
		return "", ErrDescription
	}
	return description, nil
}

var pointDecimalSeparator string = "."

const (
	cents  int64 = 1
	dolars int64 = cents * 100

	base int64 = 10000
)

type Money struct {
	whole    int64
	decimal  int64
	currency string // USD for example
}

func (m Money) toInt() int64 {
	max := []rune("0000")
	value := fmt.Sprintf("%v", m.decimal)
	for i, s := range []rune(value) {
		max[i] = s
	}
	centsString := string(max)
	cents, err := strconv.ParseInt(centsString, 10, 64)
	if err != nil {
		log.Fatalf("Could not parse int value %v", centsString)
	}
	return cents + 10000*m.whole
}

func (m Money) PreciseConvert(rate Money) Money {
	// 4 digit precision
	// .0001

	t1 := m.toInt()
	t2 := rate.toInt()

	mul := t1 * t2

	base2 := base * base
	whole := mul / base2

	decimal100 := (mul % base2) / base
	decimal := decimal100 / 100

	// if remainder, always rounds up
	if (decimal100 % 100) > 0 {
		decimal += 1
	}

	return Money{whole: whole, decimal: decimal}
}

func (m Money) ToString() string {
	return fmt.Sprintf("%v%v%v", m.whole, pointDecimalSeparator, m.decimal)
}

func (m *Money) UnmarshalJSON(b []byte) error {
	var err error
	s := string(b)
	s = s[1 : len(s)-1] // remove quotes
	*m, err = NewMoney(s)
	return err
}

func (m Money) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"%s\"", m.ToString())
	return []byte(s), nil
}

var ErrAmount = errors.New("Invalid purchase amount")

func validateAmount(amountString string) (int64, error) {
	amount, err := strconv.ParseInt(amountString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Could not parse value: %w", ErrAmount)
	}
	if amount < 0 {
		return amount, fmt.Errorf("Value should be positive: %w", ErrAmount)
	}
	return amount, nil
}

func NewMoney(valueString string) (Money, error) {

	var m Money

	// splits the value
	separatedValues := strings.Split(valueString, pointDecimalSeparator)
	if len(separatedValues) != 2 {
		return m, ErrAmount
	}
	wholeString, decimalString := separatedValues[0], separatedValues[1]

	whole, err := validateAmount(wholeString)
	if err != nil {
		return m, err
	}

	// trimm if more than 5 decimal places
	if len(decimalString) > 4 {
		decimalString = decimalString[:4]
	}

	decimal, err := validateAmount(decimalString)
	if err != nil {
		return m, err
	}

	m = Money{decimal: decimal, whole: whole, currency: "$"}

	return m, nil
}

type Transaction struct {
	Description string `json:"description"`
	Date        Time   `json:"date"`
	Amount      Money  `json:"amount"`
}

type IdentifiedTransaction struct {
	Transaction
	Uid string `json:"uid"`
}

func NewTransaction(description string, dateString string,
	value string) (Transaction, error) {

	var tr Transaction
	desc, err := NewDescription(description)
	if err != nil {
		return tr, err
	}
	amount, err := NewMoney(value)
	if err != nil {
		return tr, err
	}
	date, err := NewTime(dateString)
	if err != nil {
		return tr, err
	}
	tr = Transaction{
		Amount:      amount,
		Date:        date,
		Description: desc,
	}
	return tr, nil
}

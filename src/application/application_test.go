package application

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

type testCase struct {
	value          string
	expectedResult bool
}

func TestValidateDateFormat(t *testing.T) {
	var tests = []testCase{
		{"26/09/2023", true},
		{"26-09-2023", true},
		{"2023-09-23", true},
		{"2023/09/23", true},
		{"14/01/2020", true},

		{"01/14/2020", false},
		{"32/10/2020", false},
		{"01-14-2020", false},
		{"32-10-2020", false},
		{"2023-15-23", false},
		{"2023/15/23", false},
	}
	for _, testCase := range tests {
		t.Run(testCase.value, func(t *testing.T) {
			_, err := NewTime(testCase.value)
			if testCase.expectedResult {
				if err != nil {
					t.Errorf("Received error for valid test case (%v): %v", testCase.value, err)
				}
			} else {
				if err == nil {
					t.Errorf("No error received for invalid test case (%v)", testCase.value)
				}
				if !errors.Is(err, ErrDate) {
					t.Errorf("Error differs from expected: received (%v); expected (%v)", err, ErrDate)
				}
			}
		})
	}
}

func TestValidDescription(t *testing.T) {
	var tests = []testCase{
		{"", true},
		{"Lorem ipsum dolor sit amet, consectetur adipiscing", true},
		{"Lorem ipsum dolor sit amet, consectetur adipiscing ", false},
	}
	for _, testCase := range tests {
		t.Run(testCase.value, func(t *testing.T) {
			_, err := NewDescription(testCase.value)
			if testCase.expectedResult {
				if err != nil {
					t.Errorf("Received error for valid test case (%v): %v", testCase.value, err)
				}
			} else {
				if err == nil {
					t.Errorf("No error received for invalid test case (%v)", testCase.value)
				}
				if !errors.Is(err, ErrDescription) {
					t.Errorf("Error differs from expected: received (%v); expected (%v)", err, ErrDate)
				}
			}
		})
	}
}

func TestValidAmounts(t *testing.T) {
	var tests = []testCase{
		{"0.00", true},
		{"14990.667", true}, // taken from api call
		{"1.23", true},
		{"10.0", true},

		{"-10.20", false},
		{"", false},
		{"10.a", false},
	}

	for _, testCase := range tests {
		t.Run(testCase.value, func(t *testing.T) {
			_, err := NewMoney(testCase.value)
			if testCase.expectedResult {
				if err != nil {
					t.Errorf("Received error for valid test case (%v): %v", testCase.value, err)
				}
			} else {
				if err == nil {
					t.Errorf("No error received for invalid test case (%v)", testCase.value)
				}
				if !errors.Is(err, ErrAmount) {
					t.Errorf("Error differs from expected: received (%v); expected (%v)", err, ErrDate)
				}
			}
		})
	}

}

func TestTransactionUnmarshal(t *testing.T) {

	desc, _ := NewDescription("test")
	time, _ := NewTime("01/02/1998")
	amount, _ := NewMoney("12.34")
	expected_transaction := IdentifiedTransaction{
		Transaction{
			Description: desc,
			Date:        time,
			Amount:      amount,
		},
		"F0FE7872-E0A6-231B-13F2-2A6EBF6EF160",
	}

	input := []byte(`
	{
		"uid": "F0FE7872-E0A6-231B-13F2-2A6EBF6EF160",
		"description": "test",
		"amount": "12.34",
		"date": "01/02/1998"
	}
	`)

	var idTran IdentifiedTransaction
	if err := json.Unmarshal(input, &idTran); err != nil {
		t.Errorf("Could not unmarshal transaction %v", err)
	}

	if idTran != expected_transaction {

		t.Errorf("Expected %v, got %v", expected_transaction, idTran)
	}

}

func TestTransactionMarshal(t *testing.T) {

	desc, _ := NewDescription("test")
	time, _ := NewTime("01/02/1998")
	amount, _ := NewMoney("12.34")
	expected_transaction := IdentifiedTransaction{
		Transaction{
			Description: desc,
			Date:        time,
			Amount:      amount,
		},
		"F0FE7872-E0A6-231B-13F2-2A6EBF6EF161",
	}

	input := []byte(`
	{
		"description": "test",
		"date": "1998-02-01T00:00:00Z",
		"amount": "12.34",
		"uid": "F0FE7872-E0A6-231B-13F2-2A6EBF6EF161"
	}
	`)

	content, err := json.Marshal(expected_transaction)
	if err != nil {
		t.Errorf("Could not Marshal transaction %v", expected_transaction)
	}

	expected := make(map[string]string)
	err = json.Unmarshal(input, &expected)
	if err != nil {
		t.Errorf("Could not Unmarshal expected value %v", input)
	}

	result := make(map[string]string)
	err = json.Unmarshal(content, &result)
	if err != nil {
		t.Errorf("Could not Unmarshal result value %v", content)
	}

	eq := reflect.DeepEqual(expected, result)
	if !eq {
		t.Errorf("Expected %v, got %v", expected, result)
	}

}
func TestCreateTransaction(t *testing.T) {

	var tests = []struct {
		description    string
		date           string
		amount         string
		expectedResult bool
	}{
		{"a", "b", "c", false},
		{"b", "01/jan/1997", "0.00", false},
		{"abcdefghijklmnopqrstuvxwyzabcdefghijklmnopqrstuvxwyz", "01/04/1997", "10.90", false},

		{"valid1", "01/04/1997", "10.90", true},
	}

	for _, testCase := range tests {
		t.Run(testCase.description, func(t *testing.T) {

			_, err := NewTransaction(
				testCase.description, testCase.date, testCase.amount)

			if testCase.expectedResult {
				if err != nil {
					t.Errorf("Received error for valid test case (%v): %v", testCase.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("No error received for invalid test case (%v)", testCase.description)
				}
			}

		})
	}

}

func TestMoneyConversion(t *testing.T) {

	rate := Money{decimal: 99, whole: 12}
	value := Money{decimal: 99, whole: 12}

	newValue := value.PreciseConvert(rate)

	expected := Money{decimal: 75, whole: 168}

	if newValue.decimal != expected.decimal {
		t.Errorf("expected %v but received %v\n", expected.decimal, newValue.decimal)
	}

	if newValue.whole != expected.whole {
		t.Errorf("expected %v but received %v\n", expected.whole, newValue.whole)
	}
}

func TestMoneyTrimming(t *testing.T) {
	value, _ := NewMoney("12.12345")
	expected := Money{decimal: 1234, whole: 12}
	if value.decimal != expected.decimal {
		t.Errorf("expected %v but received %v\n", expected.decimal, value.decimal)
	}
}

func TestMoneyConsertionMoreDecimalPlaces(t *testing.T) {
	rate, _ := NewMoney("12.345")
	value, _ := NewMoney("69.788")
	newValue := rate.PreciseConvert(value)
	expected := Money{decimal: 54, whole: 861}

	if newValue.decimal != expected.decimal {
		t.Errorf("expected %v but received %v\n", expected.decimal, newValue.decimal)
	}

	if newValue.whole != expected.whole {
		t.Errorf("expected %v but received %v\n", expected.whole, newValue.whole)
	}
}

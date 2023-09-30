package application

func GetSampleIdentifiedTransaction() IdentifiedTransaction {
	return IdentifiedTransaction{
		GetSampleTransaction(),
		"F0FE7872-E0A6-231B-13F2-2A6EBF6EF160",
	}
}

func GetSampleTransaction() Transaction {
	desc, _ := NewDescription("test")
	time, _ := NewTime("01/02/1998")
	amount, _ := NewMoney("12.34")
	return Transaction{
		Description: desc,
		Date:        time,
		Amount:      amount,
	}
}

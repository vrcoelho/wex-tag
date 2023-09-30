package persistance

import (
	"encoding/json"
	"os"
	"testing"
	"time"
	"wex/src/application"
)

type testCase struct {
	description  string
	dateString   string
	amountString string
}

const testFileName = "storage_test.json"

func TestMain(m *testing.M) {
	os.Remove(testFileName)
	code := m.Run()
	os.Remove(testFileName)
	os.Exit(code)
}

func TestPersist(t *testing.T) {
	d := startDriver(testFileName)
	tran := application.GetSampleTransaction()
	uid := d.RegisterTransaction(tran)

	time.Sleep(500 * time.Millisecond)

	content, err := os.ReadFile(testFileName)
	if err != nil {
		t.Errorf("Could not read test file %v", err)

	}
	mapContent := make(map[string]application.IdentifiedTransaction)

	err = json.Unmarshal(content, &mapContent)
	if err != nil {
		t.Errorf("Could not parse test file %v", testFileName)
	}

	recordedTran, ok := mapContent[uid]
	if !ok {
		t.Errorf("Transaction not recorded in test file %v", testFileName)
	}

	expectedTransaction := application.IdentifiedTransaction{
		Transaction: tran,
		Uid:         uid,
	}

	if recordedTran != expectedTransaction {
		t.Errorf("Transaction recorded %v different from expected %v", recordedTran, expectedTransaction)
	}

}

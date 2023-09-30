package persistance

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"wex/src/application"
)

type PersistanceDriver interface {
	RegisterTransaction(application.Transaction) string
	QueryTransaction(string) (application.IdentifiedTransaction, error)
}

type Driver struct {
	mu           *sync.Mutex
	transactions map[string]application.IdentifiedTransaction
	internalFile string
	transChannel chan application.IdentifiedTransaction
}

func pseudo_uuid() (uuid string) {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

const (
	localFileName = "./../storage/localdb.json"
)

func startDriver(storageFile string) *Driver {
	d := Driver{
		internalFile: storageFile,
		transChannel: make(chan application.IdentifiedTransaction),
		mu:           &sync.Mutex{}}

	var err error
	d.transactions, err = d.loadLocalContent()
	if err != nil {
		log.Printf("Internal db (%v) not found", storageFile)
	}

	go d.monitorPersistQueue()

	return &d
}

func StartDriver() *Driver {
	return startDriver(localFileName)
}

func (d *Driver) RegisterTransaction(tran application.Transaction) string {

	var newUid string
	d.mu.Lock()
	defer d.mu.Unlock()

	for {
		newUid = pseudo_uuid()
		if _, ok := d.transactions[newUid]; !ok {
			break
		}
	}
	d.transChannel <- application.IdentifiedTransaction{Transaction: tran, Uid: newUid}
	return newUid
}

func (d *Driver) registerTransaction(tran application.IdentifiedTransaction) {

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.transactions[tran.Uid]; ok {
		log.Fatalf("Transaction %v already existed", tran.Uid)
	} else {
		d.transactions[tran.Uid] = tran
	}
}

var QueryNotFoundError = errors.New("Transaction not found")

func (d *Driver) QueryTransaction(transactionId string) (application.IdentifiedTransaction, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if transaction, ok := d.transactions[transactionId]; ok {
		return transaction, nil
	} else {
		return transaction, QueryNotFoundError
	}

}

func (d *Driver) monitorPersistQueue() {
	for {
		select {
		case newTransaction := <-d.transChannel:
			d.registerTransaction(newTransaction)
			d.persistToFile()
		}
	}

}

func (d *Driver) loadLocalContent() (map[string]application.IdentifiedTransaction, error) {

	k := make(map[string]application.IdentifiedTransaction)

	content, err := os.ReadFile(d.internalFile)
	if err != nil {
		return k, err
	}
	err = json.Unmarshal(content, &k)
	if err != nil {
		return k, err
	}
	return k, nil
}

func (d *Driver) persistToFile() {

	d.mu.Lock()
	defer d.mu.Unlock()
	content, err := json.Marshal(d.transactions)
	if err != nil {
		log.Fatalf("Could not parse internal map in memory: %v", err)
	}

	err = os.WriteFile(d.internalFile, content, 0644)
	if err != nil {
		log.Fatalf("Could not save internal db: %v", err)
	}
}

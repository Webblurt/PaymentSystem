package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type AccountStatus string

const (
	StatusActive  AccountStatus = "active"
	StatusBlocked AccountStatus = "blocked"
)

type Account struct {
	IBAN    string        `json:"iban"`
	Balance float64       `json:"balance"`
	Status  AccountStatus `json:"status"`
}

type PaymentService struct {
	accounts        map[string]*Account
	emissionIBAN    string
	destructionIBAN string
	mu              sync.Mutex
}

func NewPaymentService() *PaymentService {
	emissionIBAN := "BY00000000000000000000000000001"
	destructionIBAN := "BY0000000000000000000000000002"

	accounts := map[string]*Account{
		emissionIBAN:    {IBAN: emissionIBAN, Balance: 0, Status: StatusActive},
		destructionIBAN: {IBAN: destructionIBAN, Balance: 0, Status: StatusActive},
	}

	return &PaymentService{
		accounts:        accounts,
		emissionIBAN:    emissionIBAN,
		destructionIBAN: destructionIBAN,
	}
}

func (ps *PaymentService) GetEmissionIBAN() string {
	return ps.emissionIBAN
}

func (ps *PaymentService) GetDestructionIBAN() string {
	return ps.destructionIBAN
}

func (ps *PaymentService) CreateAccount() string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	newIBAN := GenerateIBAN()
	ps.accounts[newIBAN] = &Account{
		IBAN:    newIBAN,
		Balance: 0,
		Status:  StatusActive,
	}
	return newIBAN
}

func (ps *PaymentService) Emission(amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.accounts[ps.emissionIBAN].Balance += amount
	return nil
}

func (ps *PaymentService) DestroyMoney(fromIBAN string, amount float64) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	account, exists := ps.accounts[fromIBAN]
	if !exists {
		return fmt.Errorf("account with IBAN %s not found", fromIBAN)
	}

	if account.Status == StatusBlocked {
		return fmt.Errorf("account %s is blocked", fromIBAN)
	}

	if account.Balance < amount {
		return errors.New("insufficient funds")
	}

	account.Balance -= amount
	ps.accounts[ps.destructionIBAN].Balance += amount
	return nil
}

func (ps *PaymentService) TransferMoney(fromIBAN, toIBAN string, amount float64) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	fromAccount, existsFrom := ps.accounts[fromIBAN]
	toAccount, existsTo := ps.accounts[toIBAN]

	if !existsFrom || !existsTo {
		return errors.New("one of the accounts does not exist")
	}

	if fromAccount.Status == StatusBlocked || toAccount.Status == StatusBlocked {
		return errors.New("one of the accounts is blocked")
	}

	if fromAccount.Balance < amount {
		return errors.New("insufficient funds")
	}

	fromAccount.Balance -= amount
	toAccount.Balance += amount
	return nil
}

func (ps *PaymentService) GetAllAccounts() (string, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	accounts := make([]*Account, 0, len(ps.accounts))
	for _, acc := range ps.accounts {
		accounts = append(accounts, acc)
	}

	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func GenerateIBAN() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("BY%02dCBDC%020d", rng.Intn(100), rng.Int63n(1e18))
}

func main() {
	service := NewPaymentService()

	err := service.Emission(1000)
	if err != nil {
		fmt.Println("Error during emission:", err)
	}

	newAccount := service.CreateAccount()
	fmt.Println("New account created:", newAccount)

	err = service.TransferMoney(service.GetEmissionIBAN(), newAccount, 500)
	if err != nil {
		fmt.Println("Error during transfer:", err)
	}

	err = service.DestroyMoney(newAccount, 200)
	if err != nil {
		fmt.Println("Error during destruction:", err)
	}

	accounts, _ := service.GetAllAccounts()
	fmt.Println("All accounts:", accounts)
}

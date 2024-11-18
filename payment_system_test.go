package main

import (
	"testing"
)

func TestPaymentService(t *testing.T) {
	service := NewPaymentService()

	// Test emission
	err := service.Emission(1000)
	if err != nil {
		t.Errorf("Emission failed: %v", err)
	}

	emissionAccount := service.GetEmissionIBAN()
	if service.accounts[emissionAccount].Balance != 1000 {
		t.Errorf("Expected emission balance 1000, got %f", service.accounts[emissionAccount].Balance)
	}

	// Test creating new account
	newAccount := service.CreateAccount()
	if _, exists := service.accounts[newAccount]; !exists {
		t.Errorf("New account was not created")
	}

	// Test transferring money
	err = service.TransferMoney(emissionAccount, newAccount, 500)
	if err != nil {
		t.Errorf("Transfer failed: %v", err)
	}

	if service.accounts[newAccount].Balance != 500 {
		t.Errorf("Expected balance on new account 500, got %f", service.accounts[newAccount].Balance)
	}

	if service.accounts[emissionAccount].Balance != 500 {
		t.Errorf("Expected emission account balance 500, got %f", service.accounts[emissionAccount].Balance)
	}

	// Test insufficient funds transfer
	err = service.TransferMoney(newAccount, emissionAccount, 1000)
	if err == nil {
		t.Error("Expected error due to insufficient funds, got none")
	}

	// Test destroying money
	destructionAccount := service.GetDestructionIBAN()
	err = service.DestroyMoney(newAccount, 200)
	if err != nil {
		t.Errorf("Destroy money failed: %v", err)
	}

	if service.accounts[newAccount].Balance != 300 {
		t.Errorf("Expected balance on new account 300, got %f", service.accounts[newAccount].Balance)
	}

	if service.accounts[destructionAccount].Balance != 200 {
		t.Errorf("Expected balance on destruction account 200, got %f", service.accounts[destructionAccount].Balance)
	}

	// Test account blocking
	service.accounts[newAccount].Status = StatusBlocked
	err = service.TransferMoney(newAccount, emissionAccount, 100)
	if err == nil {
		t.Error("Expected error due to blocked account, got none")
	}
}

package wallet

import (
	"errors"

	"github.com/Yessentemir/wallet/pkg/types"
	"github.com/google/uuid"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater than 0")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrPaymentNotFound = errors.New("payment not found")

type Service struct {
	nextAccountID int64 // для генерации уникального номера аккаунта
	accounts      []*types.Account
	payments      []*types.Payment
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}

	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}

	// зачисление средств пока не рассматриваем как платеж
	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

//  FindAccountByID возвращает найденный аккаунт
func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID != accountID {
			return nil, ErrAccountNotFound
		}

		if acc.ID == accountID {
			account = acc
			break
		}
	}
	return account, nil
}

// Reject отменяет платеж и возвращает деньги на аккаунт
func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}

	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount
	return nil
}

// FindPaymentByID позволяет найти платеж по идентификатору
func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}

	return nil, ErrPaymentNotFound
}

// Repeat позволяет по идентификатору платежа повторить платеж
 func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, err := s.FindPaymentByID(paymentID) // поиск платежа по идентификатору который необходимо повторить
	if err != nil {
		return nil, err
	}
	account, err := s.FindAccountByID(payment.AccountID) // поиск аккаунта по идентификатору, который принадлежит платежу который необходимо повторить
	if err != nil {
		return nil, err
	}

	if account.Balance < payment.Amount { // проверка на наличие денег на аккаунте
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= payment.Amount // списание денег с аккаунта
	paymentID = uuid.New().String() // генерация нового идентификатора платежа
	paymentNew := &types.Payment{ // создание нового платежа
		ID:        paymentID,
		AccountID: account.ID,
		Amount:    payment.Amount,
		Category:  payment.Category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, paymentNew) // добавление платежа в список платежей
	return payment, nil
}
 



package wallet

import (
	"fmt"

	"reflect"
	"testing"

	"github.com/Yessentemir/wallet/pkg/types"
	"github.com/google/uuid"
)

func TestFindAccountById_success(t *testing.T) {
	// Создаем новый сервис
	s := Service{}
	// Создаем новый аккаунт
	account := &types.Account{
		ID:      1,
		Phone:   "+992000000001",
		Balance: 0,
	}

	// Добавляем аккаунт в сервис
	s.accounts = append(s.accounts, account)
	// Проверяем, что аккаунт найден
	if result, err := s.FindAccountByID(1); err != nil {
		t.Errorf("FindAccountById(1) = %v, %v; want nil", result, err)
	}
}

type testService struct {
	*Service // embedding(встраивание)
}

func newTestService() *testService { // функция-конструктор
	return &testService{Service: &Service{}}
}

func TestService_Reject_success(t *testing.T) {
	// создаем сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	// пробуем отменить платёж
	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}

	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't find payment by id, error = %v", err)
		return
	}
	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didn't changed, payment = %v", savedPayment)
		return
	}

	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id, error = %v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didn't changed, account = %v", savedAccount)
		return
	}

}

func TestService_FindPaymentByID_success(t *testing.T) {
	// создаем сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	// пробуем найти платёж
	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}

	// сравниваем платежи
	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
		return
	}
}

func TestService_FindPaymentByID_fail(t *testing.T) {
	// создаем сервис
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	// пробуем найти несуществующий платеж
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Error("FindPaymentByID(): must return error, returned nil")
		return
	}

	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount{
	phone:   "+992000000001",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	// регистрируем там пользователя
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}

	// пополняем его счет
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposity account, error = %v", err)
	}

	// выполняем платежи
	// можем создать слайс сразу нужной длинны, поскольку знаем размер
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		// тогда здесь работаем просто через index , а не через append
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}
	}

	return account, payments, nil
}

//  write a test that creates a payment using the Pay method and then repeats it using the Repeat method
func TestService_Repeat_success(t *testing.T) {
	// создаем сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	// пробуем повторить платеж
	payment := payments[0]
	repeatPayment, err := s.Repeat(payment.ID)
	if err != nil {
		t.Errorf("Repeat(): error = %v", err)
		return
	}

	// сравниваем платежи(ID, Account, Amount, Category, Status)
	// сравнием платежи по идентификатору ID
	if reflect.DeepEqual(payment.ID, repeatPayment.ID) {
		t.Errorf("Repeat(): payments ID's are not equal = %v", err)
		return
	}
}


func TestService_Repeat_fail(t *testing.T) {
	// создаем сервис
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
    }
	// пробуем повторить платеж
	_, err = s.Repeat(uuid.New().String())
	if err == nil {
		t.Error("Repeat(): must return error, returned nil")
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	// создаем сервис
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	// пробуем создать избранное из payment и совершить платеж из него(избранного)
	payment := payments[0]
	favoritePayment, err := s.FavoritePayment(payment.ID, "tuitionFee")
	if err != nil {
		t.Errorf("FavoritePayment(): error = %v", err)
		return
	}

	favoritePay, err := s.PayFromFavorite(favoritePayment.ID)
	if err != nil {
		t.Errorf("PayFromFavorite(): error = %v", err)
		return
	}

	// сравниваем платежи(Category)
	if !reflect.DeepEqual(payment.Category, favoritePay.Category) {
		t.Errorf("PayFromFavorite(): payments categories are not equal = %v", err)
		return
	}

    // сравниваем платежи(Amount)
	if !reflect.DeepEqual(payment.Amount, favoritePay.Amount) {
		t.Errorf("PayFromFavorite(): payments amounts are not equal = %v", err)
		return
	}

	// сравниваем платежи(AccountID)
	if !reflect.DeepEqual(payment.AccountID, favoritePay.AccountID) {
		t.Errorf("PayFromFavorite(): payments account IDs are not equal = %v", err)
		return
	}
}

func TestService_PayFromFavorite_fail(t *testing.T) {
	// создаем сервис
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	// пробуем создать избранное из payment и совершить платеж из него(избранного)
	_, err = s.PayFromFavorite(uuid.New().String())
	if err == nil {
		t.Error("PayFromFavorite(): must return error, returned nil")
		return
	}
}

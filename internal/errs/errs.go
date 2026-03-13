package errs

import (
	"errors"
	"net/http"
)

var (
	ErrInvalidStock    = errors.New("invalid stock value")
	ErrInvalidPrice    = errors.New("invalid price value")
	ErrInvalidQuantity = errors.New("invalid quantity value")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidRole     = errors.New("invalid role")
	ErrInvalidPassword = errors.New("invalid password")

	ErrCartNotFound    = errors.New("cart not found")
	ErrOrderNotFound   = errors.New("order not found")
	ErrItemNotFound    = errors.New("item not found")
	ErrProductNotFound = errors.New("product not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrTokenNotFound   = errors.New("token not found")

	ErrCartAlreadyExists  = errors.New("cart already exists")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrNotEnoughStock     = errors.New("not enough stock")
	ErrEmptyCart          = errors.New("cart is empty")
	ErrCartNotActive      = errors.New("cart is not active")
	ErrNoPermission       = errors.New("permission denied")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrVersionConflict    = errors.New("version conflict")
	ErrInvalidToken       = errors.New("invalid token")
	ErrNothingToUpdate    = errors.New("nothing to update")
	ErrMissingID          = errors.New("id is required")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrBadRequest         = errors.New("invalid request format")
	ErrOrderNotPending    = errors.New("order is not pending")
)

type HTTPError struct {
	Code    int
	Message string
}

func (h *HTTPError) Error() string {
	return h.Message
}

var errToHTTP = map[error]*HTTPError{
	ErrInvalidStock:    {Code: http.StatusBadRequest, Message: "Недопустимое значение stock"},
	ErrInvalidPrice:    {Code: http.StatusBadRequest, Message: "Недопустимая цена"},
	ErrInvalidQuantity: {Code: http.StatusBadRequest, Message: "Недопустимое количество"},
	ErrInvalidEmail:    {Code: http.StatusBadRequest, Message: "Недопустимая почта"},
	ErrInvalidRole:     {Code: http.StatusBadRequest, Message: "Недопустимая роль"},
	ErrInvalidPassword: {Code: http.StatusBadRequest, Message: "Недопустимый пароль"},
	ErrNotEnoughStock:  {Code: http.StatusBadRequest, Message: "Недостаточно товара на складе"},
	ErrEmptyCart:       {Code: http.StatusBadRequest, Message: "Корзина пуста"},
	ErrNothingToUpdate: {Code: http.StatusBadRequest, Message: "Нет данных для обновления"},
	ErrMissingID:       {Code: http.StatusBadRequest, Message: "Идентификатор обязателен"},

	ErrInvalidCredentials: {Code: http.StatusUnauthorized, Message: "Неверные учётные данные"},
	ErrInvalidToken:       {Code: http.StatusUnauthorized, Message: "Неверный токен"},
	ErrUserInactive:       {Code: http.StatusUnauthorized, Message: "Учётная запись недоступна"},

	ErrNoPermission: {Code: http.StatusForbidden, Message: "Нет разрешения"},

	ErrCartNotFound:    {Code: http.StatusNotFound, Message: "Корзина не найдена"},
	ErrOrderNotFound:   {Code: http.StatusNotFound, Message: "Заказ не найден"},
	ErrItemNotFound:    {Code: http.StatusNotFound, Message: "Элемент не найден"},
	ErrProductNotFound: {Code: http.StatusNotFound, Message: "Продукт не найден"},
	ErrUserNotFound:    {Code: http.StatusNotFound, Message: "Пользователь не найден"},
	ErrTokenNotFound:   {Code: http.StatusNotFound, Message: "Токен не найден"},

	ErrCartAlreadyExists: {Code: http.StatusConflict, Message: "Корзина уже существует"},
	ErrUserAlreadyExists: {Code: http.StatusConflict, Message: "Пользователь уже существует"},
	ErrVersionConflict:   {Code: http.StatusConflict, Message: "Конфликт версий"},

	ErrCartNotActive:   {Code: http.StatusUnprocessableEntity, Message: "Корзина неактивна"},
	ErrUnauthorized:    {Code: http.StatusUnauthorized, Message: "Неавторизован"},
	ErrBadRequest:      {Code: http.StatusBadRequest, Message: "Неверный формат запроса"},
	ErrOrderNotPending: {Code: http.StatusBadRequest, Message: "Статус уже обработан"},
}

func ToHTTPError(err error) *HTTPError {
	httpErr, ok := errToHTTP[err]
	if ok {
		return httpErr
	}
	return &HTTPError{Code: http.StatusInternalServerError, Message: "internal server error"}
}

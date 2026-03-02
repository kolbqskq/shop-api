package errs

import (
	"errors"
)

var (
	ErrCartNotFound          = errors.New("Корзина не найдена")
	ErrOrderNotFound         = errors.New("Заказ не найдена")
	ErrCartAlreadyExist      = errors.New("Корзина уже создана")
	ErrInvalidStock          = errors.New("Недопустимое значение")
	ErrInvalidPrice          = errors.New("Недопустимая цена")
	ErrInvalidQuantity       = errors.New("Недопустимое количество")
	ErrNotEnoughStock        = errors.New("Недоступное кол-во")
	ErrEmptyCart             = errors.New("Корзина пуста")
	ErrCartNotActive         = errors.New("Корзина не активна")
	ErrInvalidRemoveCartItem = errors.New("Ошибка удаления продукта из корзины")
	ErrVersionConflict       = errors.New("Ошибка сервера")
	ErrItemMissing           = errors.New("Продукт не найден")
	ErrNoPermission          = errors.New("Нет разрешения")
	ErrProductNotFound       = errors.New("Продукт не найден")
	ErrBadRequest            = errors.New("Неправильный запрос")
	ErrInvalidEmail          = errors.New("Недопустимая почта")
	ErrInvalidRole           = errors.New("Недопустимая роль")
	ErrInvalidPassword       = errors.New("Неверный пароль")
	ErrUserNotFound          = errors.New("Пользователь не найден")
	ErrUserAlreadyExists     = errors.New("Пользователь уже существует")
	ErrInvalidCredentials    = errors.New("Неверные учетные данные")
	ErrUserInactive          = errors.New("Ваша учетная запись недоступна")
)

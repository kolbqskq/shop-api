# Shop API

REST API для интернет-магазина на Go.

## Стек
- Go 1.25
- PostgreSQL 17
- Gin
- JWT авторизация
- Docker

## Запуск

1. Клонировать репозиторий
2. Создать `.env` из `.env.example`
3. Запустить:
```bash
docker-compose up -d --build
```

## API

### Auth
- `POST /auth/register` — регистрация
- `POST /auth/login` — вход
- `POST /auth/logout` — выход
- `POST /auth/refresh` — обновить токены

### Products
- `GET /products` — список продуктов
- `GET /products/:id` — продукт
- `POST /products` — создать (admin)
- `PUT /products/:id` — обновить (admin)
- `DELETE /products/:id` — удалить (admin)

### Cart
- `GET /cart` — корзина
- `POST /cart/items` — добавить товар
- `PATCH /cart/items/:id` — изменить количество
- `DELETE /cart/items/:id` — удалить товар
- `DELETE /cart/items` — очистить корзину

### Orders
- `POST /orders` — создать заказ
- `GET /orders` — список заказов
- `GET /orders/:id` — заказ
- `POST /orders/:id/pay` — оплатить

## Миграции

Применить все миграции:
```bash
go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$DATABASE_URL" up
```

Откатить все:
```bash
go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$DATABASE_URL" down-to 0
```

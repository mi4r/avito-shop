# Avito Merch Shop API

Микросервис для управления покупками мерча и транзакциями монет между сотрудниками.

## Технологии

- Go 1.23+
- PostgreSQL 16+
- Chi Router
- JWT аутентификация
- Bcrypt для хеширования паролей

## Установка и запуск

1. Клонирование репозитория:
```bash
git clone https://github.com/mi4r/avito-shop.git
cd avito-shop
```

2. Установка и обновление зависимостей
```bash
go mod tidy
```

3. Для использования API внесите ваши данные в конфиг ".env"
```bash
cp .env.bak .env # Замените значения
```

4. Миграции применяются автоматически.
Запуск миграций вручную при необходимости:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest 
```
```bash
migrate -path migrations -database postgres://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable up
```
 Откат миграции:
 ```bash
migrate -path migrations -database postgres://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable down 1
```

5. Запуск API через Docker:
```bash
docker compose up
```

## API Endpoints

### Аутентификация
```POST /api/auth```
Пример вводных данных:
```json
{
  "username": "user1",
  "password": "pass123"
}
```
Ответ:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Получение информации
```GET /api/info```
Пример вводных данных:
Загаловок
```
Authorization: Bearer <token>
```
Ответ:
```json
{
  "coins": 920,
  "inventory": [
    {"type": "t-shirt", "quantity": 1}
  ],
  "coinHistory": {
    "received": [
      {"fromUser": "user2", "amount": 50}
    ],
    "sent": [
      {"toUser": "user3", "amount": 30}
    ]
  }
}
```

### Перевод монет
```POST /api/sendCoin```
Пример вводных данных:
```json
{
  "toUser": "user2",
  "amount": 100
}
```
Ответ:  
Статус успешного выполнения или код ошибки с комментарием.

### Покупка товара
```GET /api/buy/t-shirt```
Ответ:  
Статус успешного выполнения или код ошибки с комментарием.

## Тестирование
```bash
go test ./... -coverprofile profiles/cover.out && go tool cover -func=profiles/cover.out
```

## Лицензия
```
MIT License

Copyright (c) 2025 Tiko

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

```

# Архитектура авторизации

## Слои

```
HTTP Request
    ↓
[Handler] → парсинг, валидация DTO, вызов сервиса
    ↓
[AuthService] → бизнес-логика (хеширование, проверка, выдача токена)
    ↓
[UserRepository] → доступ к БД
    ↓
PostgreSQL
```

## Потоки

### Регистрация
1. Handler принимает `{ email, password }`
2. Service проверяет уникальность email, хеширует пароль (bcrypt)
3. Repository сохраняет пользователя
4. Handler возвращает 201 + `{ "user_id": "..." }`

### Логин
1. Handler принимает `{ email, password }`
2. Service находит пользователя, проверяет пароль
3. Service генерирует JWT access token
4. Handler возвращает 200 + `{ "access_token": "...", "expires_in": 3600 }`

### OAuth (Google, Yandex, GitHub)
1. Пользователь переходит на `GET /auth/{provider}/redirect`
2. Backend редиректит на страницу провайдера (state в cookie)
3. Провайдер редиректит на `GET /auth/{provider}/callback?code=...`
4. Backend обменивает code на access token, получает userinfo
5. Создаёт пользователя (если нет) или находит по oauth_provider + oauth_provider_id
6. Выдаёт JWT и редиректит на фронтенд с `?token=...`

### Защищённый эндпоинт
1. Middleware извлекает Bearer token из заголовка
2. Middleware проверяет подпись JWT, извлекает user_id
3. Middleware кладёт user_id в context
4. Handler читает user_id из context

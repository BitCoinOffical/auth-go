1. Session-based (Cookie + Session)

Сервер хранит сессию (в памяти / Redis / DB)
Клиент получает session_id в cookie
Где используют: монолиты, SSR-приложения

2. JWT (JSON Web Token)

Stateless токен, хранится на клиенте
Access token + Refresh token
Где используют: REST API, мобильные приложения, SPA — самый популярный вариант сейчас

3. OAuth2 / OpenID Connect

Авторизация через третью сторону (Google, GitHub и т.д.)
OIDC — надстройка над OAuth2, добавляет аутентификацию
Где используют: "Войти через Google", B2C продукты

4. API Key

Статичный ключ в заголовке (X-API-Key)
Где используют: server-to-server, публичные API

5. mTLS (Mutual TLS)

Обе стороны проверяют сертификаты друг друга
Где используют: microservices, внутренняя инфраструктура


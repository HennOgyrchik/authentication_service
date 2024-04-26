## Описание
Часть сервиса аутентификации с использованием JWT-токена. 

Реализована генерация и обновление пар access-refresh токенов, возврат клиенту в формате JSON.

Refresh-токены хранятся в формате bcrypt-хеша в MongoDB.

Для шифрования access-токена применяется алгоритм SHA512.

В работе используются файлы cookie.
## Способы запуска
`make run`

или

`docker build -t medods:0.0.1 . && docker compose up`

## Описание методов
### 1) (GET) /getTokens
Генерирует и возвращает пару Access - Refresh токенов

Принимает:
```json
{
  "GUID": "b1d4ce5d-1612-3533-999c-cfa72ba94f86"
}
```

Возвращает:
```json
{
  "Access": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjMwMTcwODI4MTAzMSwiZ3VpZCI6ImIxZDRjZTVkLTE2MTItMzUzMy05OTljLWNmYTcyYmE5NGY4NiJ9.otFlj1XxCvqNyoYfsfa6wH7A8fEGHpPXcqIbFcsEtimZtIoWqNh5aACd-99mWaXle1MxBFIHTb82GQtOVttZkg",
  "Refresh": "ZjIyWXo4M1hlMHl4MTc2NDNSTjlwMTIw"
}
```
### 2) (GET) /refreshTokens
Удаляет старую пару токенов и генерирует новую.

Принимает:
```json
{
  "refresh_token": "ZjIyWXo4M1hlMHl4MTc2NDNSTjlwMTIw"
}
```

Возвращает:
```json
{
  "Access": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjMwMTcwODI4MTAzMSwiZ3VpZCI6ImIxZDRjZTVkLTE2MTItMzUzMy05OTljLWNmYTcyYmE5NGY4NiJ9.otFlj1XxCvqNyoYfsfa6wH7A8fEGHpPXcqIbFcsEtimZtIoWqNh5aACd-99mWaXle1MxBFIHTb82GQtOVttZkg",
  "Refresh": "ZjIyWXo4M1hlMHl4MTc2NDNSTjlwMTIw"
}
```
## Пример использования через curl
Внимание! Для работы приложения используются cookie.

Запрос `/getTokens`

```
curl -v --location --request GET 'http://192.168.0.116:8080/getTokens' \
--header 'Content-Type: application/json' \
--data '{
"GUID": "b1d4ce5d-1612-3533-999c-cfa72ba94f86"
}'
```

Запрос `/refreshTokens`

```
curl  --location --request GET 'http://192.168.0.116:8080/refreshTokens' \
--header 'Content-Type: application/json' \
--header 'Cookie: user=b1d4ce5d-1612-3533-999c-cfa72ba94f86' \
--data '{
"refresh_token": "MjE4STQ5WUY2MlhjNDBDMjUxMGQ5"
}'
`

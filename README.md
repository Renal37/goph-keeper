# goph-keeper


## Начала работы  
Запустить среду выполнения: `docker-compose up -d`  
Создать сертификаты: `make cert`  

## Запуск сервера  
Конфиг сервера: `./config/server.json`
```
{
  "host": "localhost:3200",
  "dsn": "postgres://user:password@localhost:5432/local?sslmode=disable",
  "jwt_key": "12345"
}
```

Переменные окружения:
```
$HOST 
$DSN
$JWT_KEY
```

Аргументы:
```
- mk "1234567812345678" ```

Пример запуска сервера:
```
go run ./cmd/server/. -mk "1234567812345678"
```

## Запуск агента  
Конфиг агента: `./config/agent.json`
```
{
  "server_addr": "localhost:3200"
}
```

Переменные окружения:
```
$JWT
```

Аргументы:
```
- c "read-file" //command for storage

Support command -c:
sign-up - create new account
sign-in - sign in with your account
read-file - read all files on your account
write-file - write file on your account
delete-file - delete file from your account
```

Пример запуска агента:
```
go run ./cmd/agent/. -c "sign-up"
```

После аутентификации пользователя следовать подсказкам на экране или добавить токен через переменные окружения `$JWT`.

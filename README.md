[![codecov](https://codecov.io/github/jbakhtin/goph-keeper/graph/badge.svg?token=i6zPByp7Tt)](https://codecov.io/github/jbakhtin/goph-keeper)

# goph-keeper
GophKeeper представляет собой клиент-серверную систему, позволяющую пользователю надёжно и безопасно хранить логины, пароли, бинарные данные и прочую приватную информацию.

![README.md Cover](docs/images/README%20image%20cover.png)

- [Техническое задание](docs/technical%20specification.md)
- [Анализ и технические требования](docs/analysis%20and%20technical%20requirements.md)
- [Проектирование приложения](docs/application%20design.md)
- [Задачи](docs/tasks.md)

## Запуск приложения 

Клонируем репозиторий на локальную машину:
```
git clone git@github.com:jbakhtin/goph-keeper.git 
```

Переключаемся на ветку разработки:
```
git fetch origin iter1-auth:iter1-auth
git checkout iter1-auth
```

### Запуск сервера

Собрать докер образы:
```
make build
```

Запустить докер образы:
```
make up
```

Сервер готов принимать запросы.

### Запуск клиента

Сборка клиента:
```
go build -o bin/client cmd/client/main.go
```

Клиент готов к работе:

Регистрация:
```
bin/client registration --email="ЭЛЕКТРОННАЯ_ПОЧТА" --password="ПАРОЛЬ" --password_confirmation="ПАРОЛЬ"
```

Авторизация:
```
bin/client login --email="ЭЛЕКТРОННАЯ_ПОЧТА" --password="ПАРОЛЬ"
```

Обновить JWT токен:
```
bin/client refreshtoken
```

Завершение сессии: (0 - текущая сессия; 1 - все сессии пользователя)
```
bin/client logout --type=[0|1] 
```
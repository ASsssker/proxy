
# Proxy

Сервис для проксировния HTTP запросов

## Запуск
```bash
    make help            # Просмотр доступных команд
    make depends         # Установка зависимостей
    make up              # Запуск сервиса через docker compose
    make migrations-up   # Применение миграций для БД
```

## Пример использования

1. Отправить запрос:
```bash
curl -X POST localhost:8080/v1/task \
        -d '{ "url": "http://google.com",
	          "method": "GET",
	          "headers": {
		          "Content-Language": "en-US"
	           },
	          "body": "Hello, world!"
            }'
# {
# 	"id": "7bb0d710-57e5-4242-968a-c79d00fa4460"
# }

```
2. Получить результаты обрабокти используя полученный id:
```bash
curl localhost:8080/v1/task/7bb0d710-57e5-4242-968a-c79d00fa4460
# {
# 	"id": "7bb0d710-57e5-4242-968a-c79d00fa4460",
#	"status": "done",
#	"http_status_code": 400,
#	"headers": {
#		"Content-Length": "1555",
#		"Content-Type": "text/html; charset=UTF-8",
#		"Date": "Sun, 11 May 2025 19:30:32 GMT",
#		"Referrer-Policy": "no-referrer"
#	},
#	"body": "<!DOCTYPE html>\n<html lang=en>\n  <meta charse...",
#	"content_length": 1555
# }
```

# My TON Storage Gateway API - Bruno Collection

Это коллекция Bruno для тестирования API My TON Storage Gateway.

## Установка и настройка

1. Установите [Bruno](https://usebruno.com/) 
2. Откройте папку `bruno-collection` как коллекцию в Bruno
3. Сконфигурируйте env(справа вверху можно создать свой енв), список необходимых переменных ниже
4. Открыть любой запрос, например на получение списка репортов `reports > Get All Reports`, нажать `Ctrl+Enter`

## Переменные среды

Обновите следующие переменные в соответствии с вашей конфигурацией:

- `base_url`: URL сервера (по умолчанию http://localhost:9093)
- `bag_id`: Тестовый ID бэга в формате hex (64 символа)
- `admin_token`: Bearer токен для административных операций
- `metrics_token`: Bearer токен для доступа к метрикам
- `reports_token`: Bearer токен для работы с жалобами
- `bans_token`: Bearer токен для работы с банами

## Структура коллекции

### Gateway Endpoints (`/api/v1/gateway`)

- **Get Bag Info** - `GET /:bagid` - Получить информацию о бэге
- **Get File or Directory** - `GET /:bagid/*` - Получить файл или содержимое директории
- **Health Check** - `GET /health` - Проверка состояния сервиса
- **Get Metrics** - `GET /metrics` - Метрики Prometheus (требует авторизации)

### Reports Endpoints (`/api/v1/reports`)

- **Get All Reports** - `GET /` - Получить все жалобы (с пагинацией)
- **Add Report** - `POST /` - Добавить новую жалобу
- **Get Reports by Bag ID** - `GET /:bagid` - Получить жалобы для конкретного бэга

### Bans Endpoints (`/api/v1/bans`)

- **Get All Bans** - `GET /` - Получить все баны (с пагинацией)
- **Update Ban Status** - `PUT /` - Обновить статус бана
- **Get Ban by Bag ID** - `GET /:bagid` - Получить информацию о бане для конкретного бэга

## Аутентификация

Большинство эндпоинтов требуют Bearer токен в заголовке Authorization:

```
Authorization: Bearer <your_token_here>
```


## Ошибки
Все ошибки возвращаются в формате:
```json
{
  "error": "error message"
}
```

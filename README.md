Инструмент для мониторинга статуса VPN соединения с использованием системного трея и Telegram-бота.

### Описание проекта

Приложение предоставляет пользователю:

1. Интеграцию с системным треем.
   - Отображает статус VPN соединения.
   - Позволяет выйти из приложения через интерфейс системного трея.

2. Инструмент для проверки доступности устройства по IP-адресу через ICMP пинг.

3. Telegram-бот для удалённой проверки статуса VPN.
   - Команда /status сообщает, находится ли VPN клиент в сети.
   - Реализована защита от слишком частых запросов.

4. Логирование событий и автоматическая ротация логов.

### Конфигурация

Программа использует конфигурационный файл config.json. Он будет автоматически создан при первом запуске приложения, если он не существует.

Пример содержимого config.json:

    {
      "vpn_client_ip": "10.9.0.2",
      "log_dir": "logs",
      "ping_interval": 10,
      "log_retention_days": 5,
      "telegram_bot_token": "ВАШ ТОКЕН",
      "ping_timeout": 10
    }

- vpn_client_ip: IP-адрес VPN клиента, который необходимо пинговать.
- log_dir: Директория, где будут храниться логи приложения.
- ping_interval: Интервал в секундах перед повторной проверкой статуса через Telegram.
- log_retention_days: Количество дней для хранения логов. Более старые файлы удаляются автоматически.
- telegram_bot_token: Токен вашего Telegram бота.
- ping_timeout: Время в секундах ожидания ответа на ICMP пинг.

### Сборка и упаковка

Для сборки исполняемого файла с включением иконки используйте следующую команду:

    go build -ldflags "-H windowsgui" -o vpn-status-monitor


### Как пользоваться

- Первый способ: отредактировать код, введя токен бота.

- Второй способ: запустить программу, первый запуск создаст конфигурационный файл config.json. Отредактироват его. 

### Установка зависимостей
1. Для библиотеки systray:   

    go get github.com/getlantern/systray   

2. Для библиотеки ping (используется для ICMP запросов):
   
   go get github.com/go-ping/ping   

3. Для библиотеки работы с Telegram Bot API:
   
   go get github.com/go-telegram-bot-api/telegram-bot-api/v5
   
Обратите внимание, что иногда библиотеки используют версии (например, /v5), убедитесь, что вы указываете правильную версию, чтобы избежать конфликтов.

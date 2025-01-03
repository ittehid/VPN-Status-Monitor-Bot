


## VPN Status Monitor Bot
Программа предназначена для мониторинга доступности VPN-клиента через Telegram-бота. Она выполняет периодические проверки доступности заданного IP-адреса с помощью пинга и отправляет уведомления в Telegram-бот о статусе клиента.

### Основные возможности:

-   Регулярная проверка доступности заданного IP-адреса через пинг.
-   Автоматическая отправка уведомлений в Telegram-бот о текущем статусе клиента каждые 30 минут.
-   Запись всех событий и ошибок в файлы логов.
-   Автоматическое удаление устаревших логов на основе заданного срока хранения.    
-   Настраиваемые параметры (IP-адрес клиента, частота пинга, тайм-ауты, параметры логирования) через файл `config.json`.

 
### Формат файла конфигурации (`config.json`):
    {
      "vpn_client_ip": "10.9.0.2",
      "log_dir": "logs",
      "ping_interval": 10,
      "log_retention_days": 5,
      "telegram_bot_token": "ВАШ_ТЕЛЕГРАМ_ТОКЕН",
      "ping_timeout": 10
    } 

### Команды Telegram-бота:
-   `/status` – проверка доступности VPN-клиента.

При первом запуске создается файл конфигурации `config.json` с настройками по умолчанию.

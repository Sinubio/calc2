![image](https://github.com/user-attachments/assets/02f93281-f428-4fb5-ba85-786dc4ce7cba)


## Требования
- Go 1.20+
- Docker (опционально)
- Переменные среды для настройки времени операций

## Запуск

Локально
1. Сервер:
```bash
cd mycalcservice
go run ./cmd/calc_service

Агент (в новом терминале):
COMPUTING_POWER=4 go run ./cmd/agent -server=http://localhost:8080

Примеры использования
1. Отправка выражения
curl -X POST -H "Content-Type: application/json" \
-d '{"expression": "3 + 5 * (2 - 8)"}' \
http://localhost:8080/api/v1/calculate


Конфигурация
TIME_ADDITION_MS
Время сложения (мс)
2000
TIME_SUBTRACTION_MS
Время вычитания (мс)
1500
TIME_MULTIPLICATION_MS
Время умножения (мс)
3000
TIME_DIVISION_MS
Время деления (мс)
5000
COMPUTING_POWER
Количество потоков агента
4





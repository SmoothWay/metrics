# Description

This is the project named  "Service of collecting and allerting metrics" from Yandex "Go advanced Developer" course
[https://practicum.yandex.ru/go-advanced/?from=catalog](link)  

There are two services: first one is server which accepts requests and stores metrics in memory, file or database. Second service collects and sends metrics to the server.

- router: Chi
- database: Postgres (pgx)
- logger: zap
- env parser: caarlos0/env/v6


## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

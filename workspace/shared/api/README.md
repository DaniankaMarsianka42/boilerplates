# Генерация api

1. создать и расписать yaml декларации
2. сбандлить декларации в один файл с помощью таски 'task redocly-cli:order-v1-bundle'
3. сгенерировать все с помощью команды '.\bin\ogen.exe --target shared\pkg\api\gen --package order --clean shared\api\bundles\order.openapi.v1.bundle.yaml'
4. команды вводить из корневого каталога workspace
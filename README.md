# Тестовое задание Avito

## Задание
Требуется реализовать сервис, хранящий пользователя и сегменты, в которых он состоит (создание, изменение, удаление сегментов, а также добавление и удаление пользователей в сегмент)

Полное описание по ссылке - [тут](https://github.com/boichique/avito-test-task/blob/main/AvitoTask.md "тут")

## Необходимые инструменты для запуска сервиса
На компьютере должны быть установлены:
- Docker (с возможностью использования docker compose)
- go

## Команды Makefile
Запуск сервиса:
- `make service-up`

Остановка сервиса:
- `make service-down`

Форматирование, проверка линтерами и прогон тестов:
- `make before-push`

## Работа с сервисом
Сервис стартует без данных, так что сначала необходимо заполнить базу пользователями и сегментами. Для запуска запросов можно использовать postman, swagger или curl. Ниже приведены примеры запросов для postman и swagger:


#### Postman:
Создание пользователя
![CreateUserPostman](https://github.com/boichique/movie-reviews/assets/87061629/e6a93895-28e9-4a7d-9109-b511584311eb)


Удаление пользователя
![DeleteUserPostman](https://github.com/boichique/movie-reviews/assets/87061629/3cccb250-d327-4d66-999a-3f6bcd37678b)


Создание сегмента
![CreateSegmentPostman](https://github.com/boichique/movie-reviews/assets/87061629/927c6219-101f-4bdb-880b-426bd3f2926f)


Удаление сегмента
![DeleteSegmentPostman](https://github.com/boichique/movie-reviews/assets/87061629/89187bdd-240b-4e2c-910e-ccffbd4efade)


Изменение сегментов пользователя
![UpdateUserSegmentsPostman](https://github.com/boichique/movie-reviews/assets/87061629/79ab26ec-50a8-43a8-bfbf-23c908af77d3)


Получение сегментов пользователя
![GetUserSegmentsPostman](https://github.com/boichique/movie-reviews/assets/87061629/58e73ef2-374e-4cce-b01f-8f5169fe74ef)


#### Swagger:
Также все запросы можно прогнать и через Swagger
![Swagger](https://github.com/boichique/movie-reviews/assets/87061629/9deba8cd-a6ba-42f2-a538-242a4342927f)

URL для подключения после запуска сервиса - [тут](http://localhost:8080/swagger/index.html "тут")

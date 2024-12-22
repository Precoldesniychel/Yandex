# Yandex
# Go Calc Service

Этот проект представляет собой веб-сервис на Go, который принимает на вход арифметическое выражение и возвращает результат его вычисления в формате JSON.

## Описание

Веб-сервис слушает по адресу `http://localhost:8080/api/v1/calculate`. Метод — **POST**.  
Формат запроса (JSON):
{
  "expression": "2+2*2"
}
Примеры запросов
Успешный запрос:
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
Ответ:
{
  "result": "6"
}
Ошибка 422 (некорректное выражение — например, содержится буква):

curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*a"
}'
Ответ:
{
  "error": "Expression is not valid"
}
Ошибка 500 (любая иная ошибка, например деление на ноль):
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "5/0"
}'
Ответ:
{
  "error": "Internal server error"
}

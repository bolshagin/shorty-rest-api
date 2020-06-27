Сокращатель ссылок
==================

Необходимо спроектировать и реализовать RESTful API для сокращателя ссылок.
Сокращатель ссылок - сервис, который позволяет пользователю создавать более
короткие адреса, которые лучше передавать другим пользователям и собирает
статистику по совершенным переходам.

**Стэк:** _MySQL, Go_

##### Работа с пользователем
- [x] Регистрация пользователя (авторизация не требуется)
- [x] Получение информации о текущем авторизированном пользователе

##### Короткие ссылки пользователя
- [x] Создание новой короткой ссылки  
- [x] Получение всех созданных коротких ссылок пользователя
- [x] Получение информации о конкретной короткой ссылке пользователя (количество переходов)
- [x] Удаление короткой ссылки пользователя

##### Статистика по ссылкам
- [ ] Получение временного графика количества переходов с группировкой по дням,
часам, минутам
- [x] Получение топа из 20 сайтов иcточников переходов

##### Ссылки
- [x] Получение редиректа на полный url по короткой ссылке

Установка 
---------
1. Клонировать репозиторий проекта
   ```sh
   $ git clone https://github.com/bolshagin/shorty-rest-api.git
   ```
2. Перейти в каталог с проектом
3. Создать необходимые для работы объекты в базе данных
    ```sql
    create table Users (
       userid int not null auto_increment primary key,
       email varchar(4000) not null,
       password varchar(4000) not null,
       access_token varchar(4000) not null
    );

    create table Links (
       linkid int not null auto_increment primary key,
       long_url varchar(8000) not null,
       short_url varchar(4000) not null default '',
       userid int not null
    );

    create table Clicks (
       linkid int not null,
       click_time datetime not null
    );
    ```
5. Сконфигурировать .toml-конфиг в папке ./configs
   ```toml
   bind_addr = ":8080"     # порт
   log_level = "debug"     # уровень логирования
   
   [store]
   dbname = "restapi_dev"  # схема бд
   user = "dev"            # пользователь
   password = "12345"      # пароль
   ```
6. С помощью makefile построить проект
   ```sh
   $ make 
   ```
   или выполнить команду
   ```sh
   $ go build -v ./cmd/apiserver 
   ```
7. Запустить получившийся бинарник
   ```
   apiserver.exe
   ```
   или
   ```
   $ ./apiserver
   ```
8. Для запусков тестов (сделал только малую часть) необходимо 
   в отдельном окне терминала (после запуска сервера) выполнить команду
   ```sh
   $ make test
   ```
   
REST API
--------
### Работа с пользователем
**/users**

`POST /users` создает пользователя с указанными в теле запроса email и паролем. 
Возвращает:
* *userid* (идентификатор пользователя в базе данных) 
* *email* (почтовый адрес пользователя)
* *access_token* (токен для дальнейшей аунтетификации запросов)

Пример запроса:
```
curl --location --request POST 'http://localhost:8080/users' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "foobar@mail.ru",
    "password": "1234567a"
}'    
```
Ответ:
```json
{
    "userid": 22,
    "email": "foobar@mail.ru",
    "access_token": "Zm9vYmFyQG1haWwucnU6MTIzNDU2N2E="
}
```
##### Коды ответов
* `201 Created` - пользователь успешно создан в базе данных
* `400 Bad request` - ошибка в формировании запроса
* `500 Internal Server Error` - ошибка возникшая при создании пользователя в базе данных (пропал коннект и т.д.)

`GET /me/` - возвращает данные по текущему авторизированному пользователю (в заголовке необходимо передать токен)
Возвращает:
* *userid* (идентификатор пользователя в базе данных) 
* *email* (почтовый адрес пользователя)

Пример запроса:
```
curl --location --request GET 'http://localhost:8080/me/' \
--header 'Authorization: Bearer Zm9vYmFyQG1haWwucnU6MTIzNDU2N2E='
```
Ответ:
```json
{
    "userid": 22,
    "email": "foobar@mail.ru"
}
```
##### Коды ответов
* `200 OK` - успешное получение данных по пользователю
* `401 Unautorized` - ошибка при аутентификации пользователя по переданному токену


### Короткие ссылки пользователя
**/links**

`POST /links` создает короткую ссылку. В теле запроса необходимо передать *long_url*. 
Ссылка будет создана для пользователя, токен которого указан в заголовке запроса.
Возвращает:
* *linkid* (идентификатор ссылки в базе данных)
* *long_url* (полная ссылка)
* *short_url* (сокращенная ссылка) 

Пример запроса:
```
curl --location --request POST 'http://localhost:8080/links' \
--header 'Authorization: Bearer Zm9vYmFyMTIzQG1haWwucnU6MTIzNDU2N2E=' \
--header 'Content-Type: application/json' \
--data-raw '{
    "long_url": "https://meduza.io/"
}'
```
Ответ:
```json
{
    "linkid": 18,
    "long_url": "https://meduza.io/",
    "short_url": "http://localhost:8080/5pKK"
}
```
##### Коды ответов
* `201 Created` - ссылка успешно создана в базе данных
* `400 Bad request` - ошибка в формировании запроса
* `401 Unautorized` - ошибка при аутентификации пользователя по переданному токену
* `500 Internal Server Error` - ошибка возникшая при создании ссылки в базе данных (пропал коннект и т.д.)

**/links/{userid}**

`GET /links/{userid}` - получает все созданные пользователем *userid* ссылки. В заголовке необходимо передать токен.
Возвращает:
* *userid* (идентификатор пользователя в базе данных) 
* *email* (почтовый адрес пользователя)
* список ссылок пользователя *links*
    * *long_url* (полная ссылка)
    * *short_url* (сокращенная ссылка)

Пример запроса:
```
curl --location --request GET 'http://localhost:8080/links/19' \
--header 'Authorization: Bearer Ym9sc2hhZ2luLm5pa2l0YUB5YW5kZXguY29tOjEyMzQ1Njc=' \
--data-raw ''
```
Ответ:
```json
{
    "userid": 19,
    "email": "foobar123@mail.ru",
    "links": [
        {
            "long_url": "https://meduza.io/",
            "short_url": "http://localhost:8080/5pKK"
        }
    ]
}
```
##### Коды ответов
* `200 OK` - успешный возврат данных
* `400 Bad request` - ошибка в формировании запроса 
* `401 Unautorized` - ошибка при аутентификации пользователя по переданному токену
* `500 Internal Server Error` - ошибка возникшая при работе с бд (пропал коннект, неверный запрос и т.д.)

`GET /link` - получает информацию по конкретной короткой ссылке *short_url* пользователя *userid*. В заголовке необходимо передать токен.
Возвращает:
* *linkid* (идентификатор ссылки в базе данных)
* *long_url* (полная ссылка)
* *short_url* (сокращенная ссылка)
* *n_clicks* (количество переходов; если переходов 0, то данного поля нет)

Пример запроса:
```
curl --location --request GET 'http://localhost:8080/link' \
--header 'Authorization: Bearer Zm9vYmFyMTIzQG1haWwucnU6MTIzNDU2N2E=' \
--header 'Content-Type: application/json' \
--data-raw '{
    "userid": 19,
    "short_url": "http://localhost:8080/5pKK"
}'
```
Ответ:
```json
{
    "linkid": 18,
    "long_url": "https://meduza.io/",
    "short_url": "http://localhost:8080/5pKK",
    "n_clicks": 1
}
```
##### Коды ответов
* `200 OK` - успешный возврат данных
* `400 Bad request` - ошибка в формировании запроса 
* `401 Unautorized` - ошибка при аутентификации пользователя по переданному токену
* `500 Internal Server Error` - ошибка возникшая при работе с бд

`DELETE /link` - удаляет короткую ссылку пользователя. Удаляет только ссылку у пользователя, токен которого передан.
Пример запроса:
```
curl --location --request DELETE 'http://localhost:8080/link' \
--header 'Authorization: Bearer Zm9vYmFyMTIzQG1haWwucnU6MTIzNDU2N2E=' \
--header 'Content-Type: application/json' \
--data-raw '{
    "short_url": "http://localhost:8080/5pKK"
}'
```
Ответ:
```json
{
    "result": "deleted"
}
```
##### Коды ответов
* `200 OK` - ссылка удалена
* `400 Bad request` - ошибка в формировании запроса 
* `401 Unautorized` - ошибка при аутентификации пользователя по переданному токену
* `500 Internal Server Error` - ошибка возникшая при работе с бд (не существует такой ссылки и т.д.)

### Статистика по ссылкам
**/stats/top**

`GET /stats/top` - возвращает топ-20 ссылок по переходам. В заголовке необходимо передать токен.
Возвращает:
* список ссылок с количеством переходов

Пример запроса:
```
curl --location --request GET 'http://localhost:8080/stats/top' \
--header 'Authorization: Bearer Zm9vYmFyMTIzQG1haWwucnU6MTIzNDU2N2E='
```

Ответ:
```json
[
   {
        "long_url": "https://meduza.io/",
        "n_clicks": 6
   }
]
```
##### Коды ответов
* `200 OK` - успешный возврат данных
* `401 Unautorized` - ошибка при аутентификации пользователя по переданному токену
* `500 Internal Server Error` - ошибка возникшая при работе с бд

### Ссылки
**/{short_url}**

`GET /{short_url}` - делает редирект по короткой ссылке

Пример запроса:
```
curl --location --request GET 'http://localhost:8080/5pKN'
```
# Тестовое задание для Avito - микросервис домов на Golang
 [Ссылка на описание задания](https://github.com/avito-tech/backend-bootcamp-assignment-2024).

---
На текущий момент готова основная часть + login + email notification.

TODO : 
* разобраться как пользоваться httptest и написать тесты ко всему.
* заменить авторизацию по id на авторизацию по email

Пока работоспособность сервиса можно проверить следующими командами:
```
app start:
    docker compose build
    docker compose up 
dummyLogin:
    curl -G -d "user_type=client"  http://localhost:8080/dummyLogin
Create house:
    curl -X POST -d "token=52855d32f4"  -d "address=sqtjt" -d "year=2000"  http://localhost:8080/house/create
Create flat: 
    curl -X POST -d "token=91a5f5db85"  -d "price=100" -d "house_id=10"  -d room=4 http://localhost:8080/flat/create   
Update flat: 
    url -X POST -d "token=91a5f5db85"  -d "price=100" -d "house_id=1" -d "id=7" -d room=40 -d "status=on moderation" http://localhost:8080/flat/update
    or
    curl -X POST -d "token=91a5f5db85"  -d "price=100" -d "house_id=1" -d "id=7" -d room=4 http://localhost:8080/flat/update
Get all flats:
    curl -G  -d "token=12e9c9f04a"     http://localhost:8080/house/1
Add user:
    curl -X POST  -d "user_type=moderator"  -d "password=123" -d "email=hhhh"   http://localhost:8080/register  
Get user token (login): 
    curl -X POST  -d "id=c05a892a-4a7d-4551-b3b0-b7f293d6050e"  -d "password=123"    http://localhost:8080/login  
Subscribe
    curl -X POST  -d "token=aca6e2bb5f"  -d "email=test@test.com"   http://localhost:8080/house/1/subscribe

```   
Для развертывания в Doker используйте соответственный config file - переименуйте confDocker.env  в conf.env

---
# Описание проделанной работы

* Функция "глупый логин" 
> - прокинул базу данных в обработчик через контекст (как "id" в словаре контекста) 
> - токен создаю через криптобезопасную функцию. Не знаю   какой размер должен быть у токена - для упрощения тестирования использую 10 символов. Храню в базе данных не сам токен, а его хэш. 
> - Видел иногда используют токены вида "Bearer 12e9c9f..." - для его использования надо раскомментить часть кода.

* HouseCreate 
> - буду передавать token в параметрах запроса.  Токен вида "79ca5f76b....464334f3"

* FlatCreate 
>  - В задании и в API есть противоречия.  задании квартиры опознаются по уникальной  комбинации дома и номера, а в API номер дома не используется совсем и квартиры распознаются по уникальному ID. Я реализовал условие из  API.
> - Для создания квартиры необходим существующий id дома - нельзя добавить квартиру к несуществующему дому или удалить дом не удаляя квартиры. 

* FlatUpdate 
 >- Также в API при обновлении квартиры ее параметры являются обязательными для запроса - нельзя просто получить текущие параметры, они всегда обновляются. Это странно, но я сделал как требовалось. 
 >- Можно обновить параметры квартиры не меняя ее статус ,  атомарно перевести ее в статус "on moderation". Или изменить статус (нельзя менять "on moderation" на "on moderation")

 * Get Flats 
 >- Получить все квартиры в доме. Ответ зависит от прав токена.
    Возвращается массив объектов квартир.
* AddUser 
>- Все поля в api не обязательные. В моей реализации поле Password и userType являются обязательными. Возвращается уникальный UUID используемый при логине. 

* Login 
>- Сравнивает хеши паролей. Удаляет старый токен из БД и выдает новый, записывая хеш токена в БД. В БД записывается дата создания токена для возможного удаления токенов с истекшим сроком использования. 

* Subscribe 
>- Возможность подписать email на обновление квартир в доме. Выполняется асинхронно при каждом изменении статуса на "approved". Выполняет 15 попыток отправить email и если не получилось жалуется в консоль7

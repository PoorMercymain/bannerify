# Тестовое задание для стажёра Backend
## Сервис баннеров
В Авито есть большое количество неоднородного контента, для которого необходимо иметь единую систему управления.  В частности, необходимо показывать разный контент пользователям в зависимости от их принадлежности к какой-либо группе. Данный контент мы будем предоставлять с помощью баннеров.
## Описание задачи
Необходимо реализовать сервис, который позволяет показывать пользователям баннеры, в зависимости от требуемой фичи и тега пользователя, а также управлять баннерами и связанными с ними тегами и фичами.
## Общие вводные
**Баннер** — это документ, описывающий какой-либо элемент пользовательского интерфейса. Технически баннер представляет собой  JSON-документ неопределенной структуры. 
**Тег** — это сущность для обозначения группы пользователей; представляет собой число (ID тега). 
**Фича** — это домен или функциональность; представляет собой число (ID фичи).  
1. Один баннер может быть связан только с одной фичей и несколькими тегами
2. При этом один тег, как и одна фича, могут принадлежать разным баннерам одновременно
3. Фича и тег однозначно определяют баннер

Так как баннеры являются для пользователя вспомогательным функционалом, допускается, если пользователь в течение короткого срока будет получать устаревшую информацию.  При этом существует часть пользователей (порядка 10%), которым обязательно получать самую актуальную информацию. Для таких пользователей нужно предусмотреть механизм получения информации напрямую из БД.
## Условия
1. Используйте этот [API](https://drive.google.com/file/d/1l4PMTPzsjksRCd_lIm0mVfh4U0Jn-A2R/view?usp=share_link)
2. Тегов и фичей небольшое количество (до 1000), RPS — 1k, SLI времени ответа — 50 мс, SLI успешности ответа — 99.99%
3. Для авторизации доступов должны использоваться 2 вида токенов: пользовательский и админский.  Получение баннера может происходить с помощью пользовательского или админского токена, а все остальные действия могут выполняться только с помощью админского токена.  
4. Реализуйте интеграционный или E2E-тест на сценарий получения баннера.
5. Если при получении баннера передан флаг use_last_revision, необходимо отдавать самую актуальную информацию.  В ином случае допускается передача информации, которая была актуальна 5 минут назад.
6. Баннеры могут быть временно выключены. Если баннер выключен, то обычные пользователи не должны его получать, при этом админы должны иметь к нему доступ.

## Дополнительные задания:
Эти задания не являются обязательными, но выполнение всех или части из них даст вам преимущество перед другими кандидатами. 
1. Адаптировать систему для значительного увеличения количества тегов и фичей, при котором допускается увеличение времени исполнения по редко запрашиваемым тегам и фичам
2. Провести нагрузочное тестирование полученного решения и приложить результаты тестирования к решению
3. Иногда получается так, что необходимо вернуться к одной из трех предыдущих версий баннера в связи с найденной ошибкой в логике, тексте и т.д.  Измените API таким образом, чтобы можно было просмотреть существующие версии баннера и выбрать подходящую версию
4. Добавить метод удаления баннеров по фиче или тегу, время ответа которого не должно превышать 100 мс, независимо от количества баннеров.  В связи с небольшим временем ответа метода, рекомендуется ознакомиться с механизмом выполнения отложенных действий 
5. Реализовать интеграционное или E2E-тестирование для остальных сценариев
6. Описать конфигурацию линтера

## Требования по стеку
- **Язык сервиса:** предпочтительным будет Go, при этом вы можете выбрать любой, удобный вам. 
- **База данных:** предпочтительной будет PostgreSQL, при этом вы можете выбрать любую, удобную вам. 
- Для **деплоя зависимостей и самого сервиса** рекомендуется использовать Docker и Docker Compose.
## Ход решения
Если у вас возникнут вопросы по заданию, ответы на которые вы не найдете в описанных «Условиях», то вы вольны принимать решения самостоятельно.  
В таком случае приложите к проекту README-файл, в котором будет список вопросов и пояснения о том, как вы решили проблему и почему именно выбранным вами способом.
## Оформление решения
Необходимо предоставить публичный git-репозиторий на любом публичном хосте (GitHub / GitLab / etc), содержащий в master/main ветке: 
1. Код сервиса
2. Makefile c командами сборки проекта / Описанная в README.md инструкция по запуску
3. Описанные в README.md вопросы/проблемы, с которыми столкнулись,  и ваша логика их решений (если требуется)

## Как запустить
Создаем `.env` файл на основе `.env.example` (или переименовываем `.env.example` в `.env`) в корне проекта, прописываем docker-compose up или make up. При конфигурации из `.env.example` сервис будет доступен на `localhost:8080`

## E2E тесты
Для e2e тестов используется отдельный набор контейнеров. Для него лучше использовать отдельный `.env`. Чтобы запустить тесты, используйте команду make test-service, или же просто команды из Makefile для test-service. Тесты проводятся над каждым эндпойнтом. Для e2e тестов указан тег сборки, и их код находится в `test/e2e/e2e_test.go`

## Конфигурация линтера
Кофигурация `golangci-lint` указана в файле `.golangci.yml`. В основном там отключены deprecated линтеры, но также отключен typecheck (но то, что ему не нравится, если бы было правдой, не позволяло бы запустить сервис), а еще tagliatelle хочет теги в стиле, противоречащем ТЗ, так что тоже отключен

## Нагрузочное тестирование
Для нагрузочного тестирования использовал K6. Скрипт на JS для этого находится в корне репозитория, называется load.js. Результаты в json-файле запаковал в архив (т.к. иначе не получилось бы залить такой большой файл). К сожалению, по максимальному времени выполнения запроса все немного грустно, но 95-ый перцентиль вроде вполне неплох.

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/db841c6f-b386-4805-87b2-0cffaabf829e)

## Эндпойнты
Подробное описание всех эндпойнтов после запуска сервиса можно найти на `localhost:8080/swagger/` (по умолчанию). Также в файле `bannerify.postman_collection.json` есть примеры запросов.

![Swagger UI](https://github.com/PoorMercymain/bannerify/assets/67076111/5c950094-fde9-4fa4-92ab-b9e1288bf149)

## Authorization
Для пользования сервисом нужно получить токен из эндпойнтов `POST /register` или `POST /acquire-token` (и помещать его при запросах к сервису в заголовок `token`). Сначала нужно зарегистрироваться через `POST /register` (чтобы зарегистрировать админа используйте заголовок `admin` со значением `true`). Там нужно придумать и ввести логин и пароль. После этого в теле ответа будет JWT токен на 1 день. 

![register](https://github.com/PoorMercymain/bannerify/assets/67076111/8bb191aa-1a32-4ad3-8b86-9ed33bf893d1)

Чтобы получить токен после регистрации, можно воспользоваться `POST /acquire-token`.

![acquire](https://github.com/PoorMercymain/bannerify/assets/67076111/8e20b227-277c-4651-afd6-3808810e5420)

## Banners
Для получения содержимого баннера пользователем используется эндпойнт `GET /user_banner`. В качестве кэша используется Redis с настройкой allkeys-lru. При задании use_last_revision=true, выдаются данные из БД (и обновляются в кэше)

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/88311c69-fbb5-4a05-b19a-0658fe1b46f8)

Для получения админом списка существующих баннеров используется эндпойнт ` GET /banner`.

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/4980e238-27b2-4664-8bc0-18db721dd9b2)

Для создания баннера используется эндпойнт `POST /banner`.

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/9b1dc49c-82c1-4fab-9fbe-69ad5b4a9034)

Для обновления баннера можно использовать эндпойнт `PATCH /banner/{id}`.

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/986cece5-f99c-4d94-8be6-32bede35e3f2)


При обновлении создается новая версия. Для того, чтобы выбрать версию, можно использовать эндпойнт `PATCH /banner_versions/choose/{banner_id}` с `version_id` в query.

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/bb890929-34d0-45ed-9094-eb421bf6ef9f)

Чтобы узнать `version_id`, нужно обратиться к списку версий баннера, доступному на `GET /banner_versions/{banner_id}`. По умолчанию он выдает до трех версий, но можно и больше, если указать в query limit больше трех.

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/c2e783b9-1dc2-4902-b50a-1c757f4f3eb4)

Для удаления баннера используется эндпойнт `DELETE /banner/{banner_id}`.

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/96a14e7d-5625-460b-947e-fc9b05f902b0)


Для удаления баннеров по tag_id или feature_id можно использовать эндпойнт `DELETE /banner` с tag_id/feature_id в query. Тут выдается 202, т.к. на сервере создается горутина для удаления (число единовременно удаляющих горутин ограничено семафором)

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/df5d17eb-b87f-402b-a6f3-8f793884d743)


Для проверки работоспособности сервиса можно использовать эндпойнт `GET /ping`.

![изображение](https://github.com/PoorMercymain/bannerify/assets/67076111/351a74fb-d630-452a-ba35-91ca440326f5)


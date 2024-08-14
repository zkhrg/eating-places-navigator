# Day 03 - Go Boot camp

## Contents

1. [Глава I](#chapter-i) \
    1.1. [General rules](#general-rules)
2. [Глава II](#chapter-ii) \
    2.1. [Rules of the day](#rules-of-the-day)
3. [Глава III](#chapter-iii) \
    3.1. [Intro](#intro)
4. [Глава IV](#chapter-iv) \
    4.1. [Упражнение 00: Загрузка данных](#exercise-00-loading-data)
5. [Глава V](#chapter-v) \
    5.1. [Упражнение 01: Simplest Interface](#exercise-01-simplest-interface)
6. [Глава VI](#chapter-vi) \
    6.1. [Упражнение 02: Proper API](#exercise-02-proper-api)
7. [Глава VII](#chapter-vii) \
    7.1. [Упражнение 03: Closest Restaurants](#exercise-03-closest-restaurants)
8. [Глава VIII](#chapter-viii) \
    8.1. [Упражнение 04: JWT](#exercise-04-jwt)


<h2 id="chapter-i">Глава I</h2>
<h2 id="general-rules">Основные правила</h2>

* Твоя программа не должна закрываться неожиданно (выдавая ошибку при корректном вводе). Если это произойдет, твой проект будет считаться неработаспособным и получит 0 во время оценки.
* Мы рекомендуем тебе писать тесты для твоего проекта, даже если если они и не оцениваются. Это даст тебе возможность легко тестировать твою работу и работу твоих пиров. Ты убедишься что тесты очень полезны, во время защиты. Во время защиты ты свободен использовать свои тесты и/или тесты пира которого ты проверяешь.
* Отправляй свою работу в нужный git репозиторий. Работа будет оцениваться только из git репозитория.
* Если твой код использует сторонние зависимости, следует использовать [Go Modules](https://go.dev/blog/using-go-modules) для управления ими.

<h2 id="chapter-ii">Глава II</h2>
<h2 id="rules-of-the-day">Правила дня</h2>

* Пиши код только в `*.go` файлах и (в случае стронних зависимостей) `go.mod` + `go.sum`
* Твой код для этого задания должен собираться с использовния простого `go build`
* Все входные данные ('page'/'lat'/'long') должны быть провалидированы и никогда не бросать HTTP 500 (только HTTP 400/401 принимаются, с понятным текстом ошибки как это объяснено EX02)

<h2 id="chapter-iii" >Глава III</h2>
<h2 id="intro" >Введение</h2>

Люди сейчас любят некоторые рекомендательные приложения. Это помогает избегать излишних дум о том что купить, куда сходить и что поесть.

Так же очень много людей имеет телефон с геолокацией. Как давно ты пытался найти рестораны в твоем районе для ужина?

Давай подумаем немного как эти сервисы работают и сделаем свой собственный, реально простой, как думаешь, сможем?

<h2 id="chapter-iv" >Глава IV</h2>
<h3 id="ex00">Упражнение 00: Загрузка данных</h3>

Очень много разных вариантов баз данных на рынке. Но потому что мы пытаемся предоставить возможность поиска, давай использовать [Elasticsearch](https://www.elastic.co/downloads/elasticsearch).

Elasticsearch это полнотекстовый поисковой движок на [Lucene](https://en.wikipedia.org/wiki/Apache_Lucene). Он предоставляет HTTP API который будет использоваться в этом задании.

Наш предоставленный набор данных о ресторанах (взято с Open Data portal) содержит более чем 13 тысяч ресторанов по всей Москве, Росси (ты можешь использовать другой похожий датасет с любой локации которой ты хочешь). Каждая запись состоит из:

```md
- ID
- Name
- Address
- Phone
- Longitude
- Latitude
```
RU:
```md
* Индивидуальный номер записи
* Название
* Адрес
* Телефон
* Долгота
* Широта
```

Перед загрузкой всех данных в базу данных, давай создадим индекс и маппинг (явное указание типов данных). Без этого Elasticsearch будет пытаться угадать типы полей основываясь на представленных данных и иногда он не может распознать геометки.

Тут парочка ссылок которые помогут тебе вкатиться в суть происходящего
* [Создание индексов](https://www.elastic.co/guide/en/elasticsearch/reference/8.4/indices-create-index.html)
* [Геометки](https://www.elastic.co/guide/en/elasticsearch/reference/8.4/geo-point.html)

Разверни по гайду из `ELASTIC.md` докер контейнер и давай немного поэксперементируем 

Для простоты давай использовать "places" как название для индекса "place" как название записи. Ты можешь создать индекс используя curl:

```sh
curl -k -XPUT -u {ваш_логин_из_ELASTIC.md}:{ваш_пароль} "https://localhost:9200/places"
```

однако в задании ты должен использвать Go Elasticsearch бинды для создания этих вещей. Следующее что тебе нужно сделать это предоставить типы для маппинга наших данных. Используя `curl` это будет выглядеть следующим образом:

```sh
curl -k -XPUT -u {ваш_логин_из_ELASTIC.md}:{ваш_пароль} https://localhost:9200/places/place/_mapping?include_type_name=true -H "Content-Type: application/json" -d @"schema.json"
```

где `schema.json` выглядит так:

```json
{
  "properties": {
    "name": {
        "type":  "text"
    },
    "address": {
        "type":  "text"
    },
    "phone": {
        "type":  "text"
    },
    "location": {
      "type": "geo_point"
    }
  }
}
```

**Важно**: приведенные `curl` команды это просто референсы для самостоятельного текстирования. Все этим действия долэны быть совершены твоей Go программой.

Сейчас твой набор данных загружен. Ты дожен использовать [Bulk API](https://www.elastic.co/guide/en/elasticsearch/reference/8.4/docs-bulk.html) для совершения этого. Все существующие Elasticsearch бинды предоставляют обертку для этого, как например [тут хороший пример](https://github.com/elastic/go-elasticsearch/blob/master/_examples/bulk/indexer.go) для официального клиента. Есть еще разные обертки, выбери любую какую захочешь.

Для проверки себя, ты можешь использовать `curl`.

```sh
curl -k -u {ваш_логин_из_ELASTIC.md}:{ваш_пароль} -s -XGET "https://localhost:9200/places?pretty"
```

ты должен получить подобный вывод:

```json
{
  "places": {
    "aliases": {},
    "mappings": {
      "properties": {
        "address": {
          "type": "text"
        },
        "id": {
          "type": "long"
        },
        "location": {
          "type": "geo_point"
        },
        "name": {
          "type": "text"
        },
        "phone": {
          "type": "text"
        }
      }
    },
    "settings": {
      "index": {
        "creation_date": "1601810777906",
        "number_of_shards": "1",
        "number_of_replicas": "1",
        "uuid": "4JKa9fgISd6-N130rpNYtQ",
        "version": {
          "created": "7090299"
        },
        "provided_name": "places"
      }
    }
  }
}
```

и запрос записи по ее ID будет выглядеть так:

```sh
curl -k -u {ваш_логин_из_ELASTIC.md}:{ваш_пароль} -s -XGET "https://localhost:9200/places/_doc/1?pretty"
```

```json
{
  "_index": "places",
  "_type": "place",
  "_id": "1",
  "_version": 1,
  "_seq_no": 0,
  "_primary_term": 1,
  "found": true,
  "_source": {
    "id": 1,
    "name": "SMETANA",
    "address": "gorod Moskva, ulitsa Egora Abakumova, dom 9",
    "phone": "(499) 183-14-10",
    "location": {
      "lat": 55.879001531303366,
      "lon": 37.71456500043604
    }
  }
}
```

Заметь, что запись с id=1 может отличаться в той что в датасете, если ты решишь использовать горутины для ускорения этого процесса. (это, однако, нет требуется в этом задании)

<h2 id="chapter-v" >Глава V</h2>
<h3 id="ex01">Упражнение 01: Простейший интерфейс</h3>

Давай создадим HTML UI (Hyper Text Markup Language, User Interface) для нашей базы данных. Скромную, мы прото должны рендерить страничку со списком названий, адресов и телефонов так чтобы пользователь мог увидеть это в браузере.

Ты должен абстрагировать твою базу данных за интерфейсом. Для простого возвращения списка записий и иметь возможность для пагинации [paginate](https://www.elastic.co/guide/en/elasticsearch/reference/current/paginate-search-results.html) по ним, этому интерфейсу достаточно:

```go
type Store interface {
    // возвращает список итемов, а так же количество совпадений и/или ошибку в случае ее возникновения
    GetPlaces(limit int, offset int) ([]types.Place, int, error)
}
```

Тут надо что бы не было связанных с Elasticsearch-ем импортов в `main` пакет, так что весь датабазовый движ должен располагаться в пакете `db` внутри твоего проекта и ты должен использовать этот интерфейс для всех взаимодействий.

Твое HTTP приложение должно работать на 8888 порту, отвечать с списком ресторанов и предоставлять простую пагинацию поверх всего. Так когда делается запрос на "http://localhost:8888?page=2" (page - это GET параметр) должна вернуться страница вида:

```html
<!doctype html>
<html>
<head>
    <meta charset="utf-8">
    <title>Places</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
</head>

<body>
<h5>Total: 13649</h5>
<ul>
    <li>
        <div>Sushi Wok</div>
        <div>gorod Moskva, prospekt Andropova, dom 30</div>
        <div>(499) 754-44-44</div>
    </li>
    <li>
        <div>Ryba i mjaso na ugljah</div>
        <div>gorod Moskva, prospekt Andropova, dom 35A</div>
        <div>(499) 612-82-69</div>
    </li>
    <li>
        <div>Hleb nasuschnyj</div>
        <div>gorod Moskva, ulitsa Arbat, dom 6/2</div>
        <div>(495) 984-91-82</div>
    </li>
    <li>
        <div>TAJJ MAHAL</div>
        <div>gorod Moskva, ulitsa Arbat, dom 6/2</div>
        <div>(495) 107-91-06</div>
    </li>
    <li>
        <div>Balalaechnaja</div>
        <div>gorod Moskva, ulitsa Arbat, dom 23, stroenie 1</div>
        <div>(905) 752-88-62</div>
    </li>
    <li>
        <div>IL Pizzaiolo</div>
        <div>gorod Moskva, ulitsa Arbat, dom 31</div>
        <div>(495) 933-28-34</div>
    </li>
    <li>
        <div>Bufet pri Astrahanskih banjah</div>
        <div>gorod Moskva, Astrahanskij pereulok, dom 5/9</div>
        <div>(495) 344-11-68</div>
    </li>
    <li>
        <div>MU-MU</div>
        <div>gorod Moskva, Baumanskaja ulitsa, dom 35/1</div>
        <div>(499) 261-33-58</div>
    </li>
    <li>
        <div>Bek tu Blek</div>
        <div>gorod Moskva, Tatarskaja ulitsa, dom 14</div>
        <div>(495) 916-90-55</div>
    </li>
    <li>
        <div>Glav Pirog</div>
        <div>gorod Moskva, Begovaja ulitsa, dom 17, korpus 1</div>
        <div>(926) 554-54-08</div>
    </li>
</ul>
<a href="/?page=1">Previous</a>
<a href="/?page=3">Next</a>
<a href="/?page=1364">Last</a>
</body>
</html>
```

"Previous" ссылка должна исчезать на страничке 1 и "Next" ссылка должна исчезать на последней странице.

**Важно**: Предупреждаю тебя что по умолчанию Elasticsearch не может предоставлять пагинацию для более чем 10000 записей. Тут есть два пути для обхода:
* Использовать Scroll API
* Или просто повысить лимит в настройках индекса специально для этого задания.
Последний вариант подходит для этого задания, однако это не лучший способ для использования в продакшене. Запрос для изменения настроек ниже:

```sh
~$ curl -XPUT -H "Content-Type: application/json" "http://localhost:9200/places/_settings" -d '
{
  "index" : {
    "max_result_window" : 20000
  }
}'
```

Также, в случае если `page` параметр задан неправильным значением (вне [0..последняя страница] или не числом) твоя страница должна возвращать 400 ошибку и обычный текст с описанием проблемы:

```txt
Invalid 'page' value: 'foo'
```

<h2 id="chapter-vi" >Глава VI</h2>
<h3 id="ex02">Упражнение 02: Собственный API</h3>

В современно мире большениство приложений предпочитают API поверх простого HTML. Так что в этом упражнении все что тебе нужно сделать это имплементировать обработчик, который будт отвечать с `Content-Type: application/json` и JSON версией на ту же самую штуку как в Уп01 (например http://127.0.0.1:8888/api/places?page=3):

```json
{
  "name": "Places",
  "total": 13649,
  "places": [
    {
      "id": 65,
      "name": "AZERBAJDZhAN",
      "address": "gorod Moskva, ulitsa Dem'jana Bednogo, dom 4",
      "phone": "(495) 946-34-30",
      "location": {
        "lat": 55.769830485601204,
        "lon": 37.486914061171504
      }
    },
    {
      "id": 69,
      "name": "Vojazh",
      "address": "gorod Moskva, Beskudnikovskij bul'var, dom 57, korpus 1",
      "phone": "(499) 485-20-00",
      "location": {
        "lat": 55.872553383512496,
        "lon": 37.538326789741
      }
    },
    {
      "id": 70,
      "name": "GBOU Shkola № 1411 (267)",
      "address": "gorod Moskva, ulitsa Bestuzhevyh, dom 23",
      "phone": "(499) 404-15-09",
      "location": {
        "lat": 55.87213179130298,
        "lon": 37.609625999999984
      }
    },
    {
      "id": 71,
      "name": "Zhigulevskoe",
      "address": "gorod Moskva, Bibirevskaja ulitsa, dom 7, korpus 1",
      "phone": "(964) 565-61-28",
      "location": {
        "lat": 55.88024342230735,
        "lon": 37.59308635976602
      }
    },
    {
      "id": 75,
      "name": "Hinkal'naja",
      "address": "gorod Moskva, ulitsa Marshala Birjuzova, dom 16",
      "phone": "(499) 728-47-01",
      "location": {
        "lat": 55.79476126986192,
        "lon": 37.491709793339744
      }
    },
    {
      "id": 76,
      "name": "ShAURMA ZhI",
      "address": "gorod Moskva, ulitsa Marshala Birjuzova, dom 19",
      "phone": "(903) 018-74-64",
      "location": {
        "lat": 55.794378830665885,
        "lon": 37.49112002224252
      }
    },
    {
      "id": 80,
      "name": "Bufet Shkola № 554",
      "address": "gorod Moskva, Bolotnikovskaja ulitsa, dom 47, korpus 1",
      "phone": "(929) 623-03-21",
      "location": {
        "lat": 55.66186417434049,
        "lon": 37.58323602169326
      }
    },
    {
      "id": 83,
      "name": "Kafe",
      "address": "gorod Moskva, 1-j Botkinskij proezd, dom 2/6",
      "phone": "(495) 945-22-34",
      "location": {
        "lat": 55.781141341601696,
        "lon": 37.55643137063551
      }
    },
    {
      "id": 84,
      "name": "STARYJ BATUM'",
      "address": "gorod Moskva, ulitsa Akademika Bochvara, dom 7, korpus 1",
      "phone": "(495) 942-44-85",
      "location": {
        "lat": 55.8060307318284,
        "lon": 37.461669109923506
      }
    },
    {
      "id": 89,
      "name": "Cheburechnaja SSSR",
      "address": "gorod Moskva, Bol'shaja Bronnaja ulitsa, dom 27/4",
      "phone": "(495) 694-54-76",
      "location": {
        "lat": 55.764134959774346,
        "lon": 37.60256453956346
      }
    }
  ],
  "prev_page": 2,
  "next_page": 4,
  "last_page": 1364
}
```

Также, в случае если `page` параметр задан неправильным значением (вне [0..последняя страница] или не числом) твой API должен возвращать 400 ошибку и JSON следующего вида:


```json
{
    "error": "Invalid 'page' value: 'foo'"
}
```

<h2 id="chapter-vii" >Глава VII</h2>
<h3 id="ex03">Упражнение 03: Ближайшие рестораны</h3>

Сейчас давай имплементируем основной кусок функционала - поиск **трех** ближайших ресторанов. Для того чтобы это сделать ты должен настроить сортировку для своего запроса:

```json
"sort": [
    {
      "_geo_distance": {
        "location": {
          "lat": 55.674,
          "lon": 37.666
        },
        "order": "asc",
        "unit": "km",
        "mode": "min",
        "distance_type": "arc",
        "ignore_unmapped": true
      }
    }
]
```

где "lat" и "lon" это твои текущие координаты. Так для http://127.0.0.1:8888/api/recommend?lat=55.674&lon=37.666 твое приложение должна вернуть JSON вида:

```json
{
  "name": "Recommendation",
  "places": [
    {
      "id": 30,
      "name": "Ryba i mjaso na ugljah",
      "address": "gorod Moskva, prospekt Andropova, dom 35A",
      "phone": "(499) 612-82-69",
      "location": {
        "lat": 55.67396575768212,
        "lon": 37.66626689310591
      }
    },
    {
      "id": 3348,
      "name": "Pizzamento",
      "address": "gorod Moskva, prospekt Andropova, dom 37",
      "phone": "(499) 612-33-88",
      "location": {
        "lat": 55.673075576456,
        "lon": 37.664533747576
      }
    },
    {
      "id": 3347,
      "name": "KOFEJNJa «KAPUChINOFF»",
      "address": "gorod Moskva, prospekt Andropova, dom 37",
      "phone": "(499) 612-33-88",
      "location": {
        "lat": 55.672865251005106,
        "lon": 37.6645689561318
      }
    }
  ]
}
```

<h2 id="chapter-viii" >Глава VIII</h2>
<h3 id="ex04">Упражнение 04: JWT</h3>

Так, последнее (не но не менее важное) что мы должны сделать это предоставить какую-нибудь простую форму аутентификации. Сейчас один из популярнейших путей это реализовать это для API используя [JWT](https://jwt.io/introduction/). К счастью, Go имеет отличный набор инструментов для того чтобы с этим справиться.

Первое что тебе нужно сделать это API endpoint http://127.0.0.1:8888/api/get_token преднозначение которого будет генерировать токен и возвращать его (как в примере)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjoxNjAxOTc1ODI5LCJuYW1lIjoiTmlrb2xheSJ9.FqsRe0t9YhvEC3hK1pCWumGvrJgz9k9WvhJgO8HsIa8"
}
```

Не забудь задать хэдер 'Content-Type: application/json'

Второе что тебе нужно сделать это защитить `/api/recommend` эндпоинт с JWT middleware, который будет проверять валидность токена.

Так, по умолчанию когда ты запрашиваешь апи из браузере он дожен фейлить с HTTP 401 ошибкой, но работать когда `Authorization: Bearer <token>` заголовок задан клиентом (ты можешь для проверок использовать `curl` или [Postman](https://www.postman.com/))

Это простейший путь для предоставления функционала аутентификации, не нужно углубляться в детали на данном этапе.



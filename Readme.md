# Злобный брат-близнец PrometheusDB

## Запуск
```shell
docker compose up # docker-compose up 
```

## Доступные ендпоинты (примеры в try.sh )
1) GET /health
2) GET /series
3) GET /query
4) POST /write


## [https://docs.google.com/document/d/11OfJM226jPn12n8kMkyefUUKAimqHwyJkqLlkwfKo4I/edit?usp=sharing](Лаба ХАСД)
! Чтобы запустить benchmark, надо добавить зависимости protobuf - с текущим go.mod не пойдет) 
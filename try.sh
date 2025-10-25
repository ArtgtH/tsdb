#!/bin/bash
# работаем
curl "http://localhost:8080/health"
# добавляем ряд/данные в ряд
curl -X POST http://localhost:8080/write   -H "Content-Type: application/json"   -d '{
    "series": [
      {
        "metric": "GPU",
        "tags": {
          "org": "test-1",
          "env": "prod",
          "server": "ND-1234"
        },
        "points": [
          {"timestamp": 1609459200000000000, "value": 15.2},
          {"timestamp": 1609459260000000000, "value": 16.8},
          {"timestamp": 1609459320000000000, "value": 14.5}
       ]
      }
    ]
  }'
# все ряды
curl "http://localhost:8080/series"
# получить конкретную метрику
curl "http://localhost:8080/query?metric=GPU&start=0&end=1700000000000000000&server=ND-1234"

# pr-review-service
## Запуск
Запустить приложение можно с помощью
```
docker-compose up
```

## Модульное тестирование
Модульное тестирование можно запустить с помощью
```
make generate-mocks
make test-unit
```

## E2E тестирование
E2E тестирование можно запустить с помощью
```
make test-e2e
```

## Нагрузочное тестирование
Нагрузочное тестирование было проведено с использованием k6 и сценария, лежащего в [load.js](https://github.com/111zxc/pr-review-service/blob/main/load.js)
```
  █ TOTAL RESULTS

    checks_total.......: 9084    300.254069/s
    checks_succeeded...: 100.00% 9084 out of 9084
    checks_failed......: 0.00%   0 out of 9084

    ✓ team created
    ✓ pr created
    ✓ reassign ok
    ✓ merge ok

    HTTP
    http_req_duration..............: avg=15.73ms  min=1.57ms   med=15.63ms  max=115.52ms p(90)=23.62ms p(95)=27.15ms
      { expected_response:true }...: avg=17.55ms  min=7.86ms   med=16.55ms  max=115.52ms p(90)=24.29ms p(95)=28ms
    http_req_failed................: 12.64% 1149 out of 9084
    http_reqs......................: 9084   300.254069/s

    EXECUTION
    iteration_duration.............: avg=265.31ms min=240.99ms med=262.65ms max=387.42ms p(90)=282.3ms p(95)=291.91ms
    iterations.....................: 2271   75.063517/s
    vus............................: 20     min=20           max=20
    vus_max........................: 20     min=20           max=20

    NETWORK
    data_received..................: 4.7 MB 157 kB/s
    data_sent......................: 3.6 MB 120 kB/s




running (0m30.3s), 00/20 VUs, 2271 complete and 0 interrupted iterations
default ✓ [======================================] 20 VUs  30s
```

## Линтеры
В проекте используются govet, staticcheck, ineffassign, unused, gosimple,
typecheck, errcheck, gocyclo, dupl, revive,
stylecheck, misspell, whitespace. Это покрывает ошибки стиля, ошибки типов, цикломатику, дублирование и другое
Конфигурация лежит в [.golangci.yml](https://github.com/111zxc/pr-review-service/blob/main/.golangci.yml)

## Эндпоинт статистики
Эндпоинт статистики доступен по GET /stats и возвращает JSON вида
```
{
    "event_counts": {
        "pr_created": 0,
        "pr_merged": 0,
        "reviewer_assigned": 0,
        "reviewer_reassigned": 0,
        "reviewer_unassigned": 0
    },
    "total_events": 0
}
```

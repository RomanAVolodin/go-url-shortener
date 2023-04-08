# go-musthave-shortener-tpl
![Coverage](https://img.shields.io/badge/Coverage-67.0%25-yellow)

### Run app 

```bash
go run cmd/shortener/main.go -a localhost:8080 -b http://localhost:8080 -f storage.json -d postgres://shortener:secret@localhost:5432/shortener
```


### Test increment 13

```bash
./cmd/shortener/shortenertest -test.v -test.run=^TestIteration13$ -binary-path=cmd/shortener/shortener -source-path=./ -database-dsn=postgres://shortener:secret@localhost:5432/shortener
```

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/shortener/proto/shortener.proto
```
Шаблон репозитория для практического трек "Веб-разработка на Go"

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` - адрес вашего репозитория на Github без
   префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона выполните следующую команды:

```
git remote add -m main template https://github.com/yandex-praktikum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

затем добавьте полученые изменения в свой репозиторий.

# Запуск автотестов

Для успешного запуска автотестов вам необходимо давать вашим веткам названия вида `iter<number>`, где `<number>` -
порядковый номер итерации.

Например в ветке с названием `iter4` запустятся автотесты для итераций с первой по четвертую.

При мерже ветки с итерацией в основную ветку (`main`) будут запускаться все автотесты.
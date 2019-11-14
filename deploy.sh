GOOS=linux GOARCH=amd64 go build -o ./blocks-stream-receiver ./cmd/blocks-stream-receiver/ && \
  scp ./blocks-stream-receiver root@144.76.140.152:/tmp/blocks-stream-receiver && \
  rm ./blocks-stream-receiver

GOOS=linux GOARCH=amd64 go build -o ./ton-api ./cmd/ton-api/ && \
  scp ./ton-api root@144.76.140.152:/tmp/ton-api && \
  rm ./ton-api

GOOS=linux GOARCH=amd64 go build -o ./ton-api-site ./cmd/ton-api-site/ && \
  scp ./ton-api-site root@144.76.140.152:/tmp/ton-api-site && \
  rm ./ton-api-site


scp yakud@95.216.33.209:/home/yakud/swagger/swagger-linux /tmp/swagger-linux
scp /tmp/swagger-linux root@144.76.140.152:/tmp/swagger-linux
scp ./swagger/swagger.yml root@144.76.140.152:/tmp/ton-api-swagger.yml

nohup /tmp/swagger-linux serve \
    --host=192.168.100.3 \
    --port=51867 \
    --flavor=redoc \
    /tmp/ton-api-swagger.yml \
    --no-open &

#  scp ./blocks-stream-receiver akisilev@46.4.4.150:/tmp/blocks-stream-receiver && \

# BUILD:
В tondb нужно сбилдить:
./cmd/blocks-stream-receiver/
./cmd/ton-api/

в ton-fork билдим:
validator-engine
validator-engine-console
lite-client
generate-random-id
blocks-stream-reader

Как запускать:
------------------------------------------------------------------------
Сначала запускается validator-engine стандартным способом.
При запуске нужно добавить параметр: --streamfile "/data/ton/ton-stream/blocks.log".
При успешном запуске validator-engine будет созван файл blocks.log и рядом с ним blocks.log.index
Как пример запуска:
validator-engine \
    -C /FOLDER/ton-work/db/etc/ton-global.config.json \
    --db /data/ton/ton-work/db \
    --ip IP:PORT \
    -l /FOLDER/ton-work/log/ton.log \
    -t 8 \
    --streamfile "/FOLDER/ton-stream/blocks.log"
------------------------------------------------------------------------
Далее запускается blocks-stream-receiver. Ожидает ENV переменные:
ADDR=0.0.0.0:7315 - адрес для открытия TCP сокета (порт должен быть доступен для blocks-stream-reader)
PROM_ADDR=0.0.0.0:8080 - prometheus endpoint (сейчас там ничего интересного, на будущее)
CH_ADDR=http://user:pass@127.0.0.1:8123/default?max_query_size=3145728000 - строка подключения к Clickhouse

После успешного запуска, в clickhouse будут созданы несколько табличек.
Для остановки - посылать SIGTERM.
------------------------------------------------------------------------
После, запускаем blocks-stream-reader. Пример:

blocks-stream-reader \
  --streamfile /FOLDER/ton-stream/blocks.log  \
  --indexfile /FOLDER/ton-stream/blocks.log.index \
  --host "0.0.0.0"\
  --port 7315 \
  --workers 3

После успешного запуска начинается стриминт.
Будет создан файл /FOLDER/ton-stream/blocks.log.index.seek в нем хранится смещение последней прочитанной строчки индекса.
Служит для возобновления работы стрима.
Для остановки - посылать SIGTERM.
------------------------------------------------------------------------
Запуск ton-api.
Ожидает ENV переменные:
ADDR=0.0.0.0:8512 - rest endpoint
CH_ADDR=http://user:pass@127.0.0.1:8123/default?max_query_size=3145728000 - строка подключения к Clickhouse
------------------------------------------------------------------------
Алгоритм остановки строго по очереди:
1. Посылаем SIGTERM в validator-engine, дожидаемся остановки
2. Посылаем SIGTERM в blocks-stream-reader, дожидаемся остановки
3. Посылаем SIGTERM в blocks-stream-receiver, дожидаемся остановки
4. Посылаем SIGTERM в ton-api, дожидаемся остановки


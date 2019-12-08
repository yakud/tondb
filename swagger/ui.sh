

nohup /tmp/swagger-linux serve \
    --host=192.168.100.3 \
    --port=51867 \
    --flavor=redoc \
    /tmp/ton-api-swagger.yml \
    --no-open > /var/log/ton-api-swagger-ui.log &

nohup /tmp/swagger-linux serve \
    --host=192.168.100.3 \
    --port=51877 \
    --flavor=redoc \
    /tmp/ton-api-public-swagger.yml \
    --no-open > /var/log/ton-api-public-swagger-ui.log &

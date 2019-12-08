
nohup /tmp/tmp.aTNqJUNFUq/build/validator-engine/validator-engine \
    -C /data/ton/ton-work/db/etc/ton-global.config.json \
    --db /data/ton/ton-work/db \
    --ip 144.76.140.152:8269 \
    -l /data/ton/ton-work/log/ton.log \
    -t 8 \
    --streamfile "/data/ton/ton-stream/blocks.log" \
    >> /data/ton/log/validator.log &

tail -f /data/ton/log/validator.log

CH_ADDR=http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton2?max_query_size=3145728000 \
  ADDR=127.0.0.1:7315 \
  PROM_ADDR=127.0.0.1:17315 \
  nohup /tmp/blocks-stream-receiver >> /data/ton/log/stream-receiver.log &

nohup /tmp/tmp.aTNqJUNFUq/build/blocks-stream/blocks-stream-reader \
  --streamfile /data/ton/ton-stream/blocks.log \
  --indexfile /data/ton/ton-stream/blocks.log.index \
  --host "127.0.0.1" \
  --port 7315 \
  --workers 3 \
  >> /data/ton/log/stream-reader.log &

CH_ADDR=http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton2?max_query_size=3145728000 \
  ADDR=192.168.100.3:8512 \
  nohup /tmp/ton-api >> /data/ton/log/api.log &

CH_ADDR=http://default:V9AQZJFNX4ygj2vP@192.168.100.3:8123/ton2?max_query_size=3145728000 \
  ADDR=192.168.100.3:8513 \
  nohup /tmp/ton-api-site >> /data/ton/log/api-site.log &

curl -XGET 'http://144.76.140.152:8512/workchain/block/masterchain?shard_hex=e000000000000000&seq_no=612162'
time curl -u tonapi:QnrWW9q4XVt5fGCcaNGvkNfQ -XGET 'http://144.76.140.152:8512/height/synced'
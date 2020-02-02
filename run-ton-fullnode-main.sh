
nohup /data/ton/bin/validator-engine \
    -C /data/ton/ton-work/db/etc/ton-global.config.json \
    --db /data/ton/ton-work/db \
    --ip 95.217.92.176:8269 \
    -l /data/ton/ton-work/log/ton.log \
    -t 8 \
    --streamblocksfile "/data/ton/ton-stream/blocks.log" \
    --streamstatefile "/data/ton/ton-stream/state.log" \
    >> /data/ton/log/validator.log &

tail -f /data/ton/log/validator.log

CH_ADDR=http://clickhouse01.pay-mills.loc:8123/default?max_query_size=3145728000 \
  ADDR=127.0.0.1:7315 \
  PROM_ADDR=127.0.0.1:17315 \
  nohup /data/ton/bin/blocks-stream-receiver >> /data/ton/log/stream-receiver.log &

nohup /data/ton/bin/blocks-stream-reader \
  --streamblocksfile /data/ton/ton-stream/blocks.log \
  --streamblocksindexfile /data/ton/ton-stream/blocks.log.index \
  --streamstatefile /data/ton/ton-stream/state.log \
  --streamstateindexfile /data/ton/ton-stream/state.log.index \
  --host "127.0.0.1" \
  --port 7315 \
  --workers 3 \
  >> /data/ton/log/stream-reader.log &

CH_ADDR=http://clickhouse01.pay-mills.loc:8123/default?max_query_size=3145728000 \
  ADDR=0.0.0.0:8512 \
  nohup /data/ton/bin/ton-api >> /data/ton/log/api.log &

curl -XGET 'http://144.76.140.152:8512/workchain/block/masterchain?shard_hex=e000000000000000&seq_no=612162'
time curl -u tonapi:QnrWW9q4XVt5fGCcaNGvkNfQ -XGET 'http://144.76.140.152:8512/height/synced'
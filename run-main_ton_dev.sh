#!/bin/bash

CH_ADDR=http://10.236.0.5:8123/main_ton_dev?max_query_size=3145728000 \
  ADDR=127.0.0.1:8316 \
  PROM_ADDR=127.0.0.1:18316 \
  nohup /data/ton/bin/blocks-stream-receiver >> /data/ton-work-net.ton.dev/log/blocks-stream-receiver.log &

nohup /data/ton/bin/blocks-stream-reader \
  --streamblocksfile /data/ton-work-net.ton.dev/stream/blocks.log \
  --streamblocksindexfile /data/ton-work-net.ton.dev/stream/blocks.log.index \
  --streamstatefile /data/ton-work-net.ton.dev/stream/state.log \
  --streamstateindexfile /data/ton-work-net.ton.dev/stream/state.log.index \
  --host "127.0.0.1"\
  --port 8316 \
  --workers 2 \
  >> /data/ton-work-net.ton.dev/log/blocks-stream-reader.log &


# TMP DEBUG
nohup /tmp/tmp.aTNqJUNFUq/build/validator-engine/validator-engine \
    -C /data/ton/ton-work/db/etc/ton-global.config.json \
    --db /data/ton/ton-work/db \
    --ip 144.76.140.152:8269 \
    -l /data/ton/ton-work/log/ton.log \
    -t 8 \
    --streamblocksfile "/data/ton/ton-stream/blocks.log" \
    --streamstatefile "/data/ton/ton-stream/state.log" \
    >> /data/ton/log/validator.log &



nohup /data/ton/bin/validator-engine \
  -C /data/ton/ton-work.old/db/etc/ton-global.config.json \
  --db /data/ton/ton-work.old/db \
  --ip 144.76.140.152:8268 \
   -l /data/ton/ton-work.old/log/ton.log \
   -t 4   --streamblocksfile "/data/ton/ton-work.old/ton-stream/blocks.log" \
   --streamstatefile "/data/ton/ton-work.old/ton-stream/state.log" \
   --streamfile "/data/ton/ton-work.old/ton-stream/stream.log" \
   >> /data/ton/ton-work.old/log/validator-engine.log &


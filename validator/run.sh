#!/bin/bash

nohup /data/ton/src/build/validator-engine/validator-engine \
  -C /data/ton/ton-work-validator/db/etc/ton-global.config.json \
  --db /data/ton/ton-work-validator/db \
  --ip 144.76.140.152:18269 \
  -l /data/ton/ton-work-validator/log/ton.log \
  -t 8 \
  >> /data/ton/log/validator-main.log &
#!/bin/bash


"${TON_BUILD_DIR}/crypto/fift" -I "${TON_SRC_DIR}/crypto/fift/lib:${TON_SRC_DIR}/crypto/smartcont" -s new-wallet.fif -1 "validator_2"

xxd -p validator_2.pk

"${TON_BUILD_DIR}/lite-client/lite-client" --verbosity 9 -p "${KEYS_DIR}/liteserver.pub" -a 127.0.0.1:3031 -rc "sendfile validator_2-query.boc" -rc "quit"


"${UTILS_DIR}/tonos-cli" call "${MSIG_ADDR}" submitTransaction \
  "{\"dest\":\"-1:2f5d8b71accd31330da5c21da11cb9947b640607960450124501eeadb21e80d7\",\"value\":\"25000000000000\",\"bounce\":false,\"allBalance\":false,\"payload\":\"\"}" \
  --abi "${CONFIGS_DIR}/SafeMultisigWallet.abi.json" \
  --sign "${KEYS_DIR}/msig.keys.json"

tonos-cli call "-1:8b2f47066d5c00320163064d2af2810637e6dc9a7cc08992a2e236e7ecce289b" \
  confirmTransaction \
  "{\"transactionId\":\"0x5eb6ea1e70eaa241\"}" \
  --abi "${CONFIGS_DIR}/SafeMultisigWallet.abi.json" \
  --sign "/home/akiselev/ton-keys/key1/msig.keys.json"


"${TON_BUILD_DIR}/crypto/fift" -I "${TON_SRC_DIR}/crypto/fift/lib:${TON_SRC_DIR}/crypto/smartcont" -s show-addr.fif  "validator_2"

tonos-cli deploy \
  /data/net.ton.dev/configs/SafeMultisigWallet.tvc \
  "{\"owners\":[\"0xaa053b4d0503884eb66915f908193f3f94bf9f526fa64f588c6562a1aa824cf2\",\"0x264d1b03503a0a5bc00d64b50232092d6f5ccdc5f137af0f821400464a1a509b\",\"0xf7cdee2eb4dda42a297285b3165515afb56a6073b941f3b8a99b2bd932e8e61c\",\"0x091f94d7b65ce05dff3994ef5598f92a188ec3cb3a34721432e68b873920bf13\",\"0xd272081bfc1d6418b1483559d5e3358e1b5108eb0bfec3b00b570b8605c4041b\"],\"reqConfirms\":4}" \
  --abi /data/net.ton.dev/configs/SafeMultisigWallet.abi.json \
  --sign multisig.wallet.keys.json \
  --wc -1


#tonos-cli call -1:2f5d8b71accd31330da5c21da11cb9947b640607960450124501eeadb21e80d7 \
#  constructor "{\"owners\":[\"0xaa053b4d0503884eb66915f908193f3f94bf9f526fa64f588c6562a1aa824cf2\",\"0x264d1b03503a0a5bc00d64b50232092d6f5ccdc5f137af0f821400464a1a509b\",\"0xf7cdee2eb4dda42a297285b3165515afb56a6073b941f3b8a99b2bd932e8e61c\",\"0x091f94d7b65ce05dff3994ef5598f92a188ec3cb3a34721432e68b873920bf13\",\"0xd272081bfc1d6418b1483559d5e3358e1b5108eb0bfec3b00b570b8605c4041b\"],\"reqConfirms\":4}" \
#  --abi /data/net.ton.dev/configs/SafeMultisigWallet.abi.json \
#  --sign multisig.wallet.keys.json
#
#
#"${TON_BUILD_DIR}/crypto/fift" -I "${TON_SRC_DIR}/crypto/fift/lib:${TON_SRC_DIR}/crypto/smartcont" -s wallet.fif validator_2 -1:8b2f47066d5c00320163064d2af2810637e6dc9a7cc08992a2e236e7ecce289b 1 24999
#
#
#
#fift -I<source-directory>/crypto/fift/lib:<source-directory>/crypto/smartcont -s wallet.fif <your-wallet-id> <destination-addr> <your-wallet-seqno> <gram-amount>


tonos-cli genaddr SafeMultisigWallet.tvc --genkey contract_keys.json

tonos-cli call "-1:MY_ADDR" submitTransaction \
  "{\"dest\":\"NEW_ADDR\",\"value\":\"10000000000\",\"bounce\":false,\"allBalance\":false,\"payload\":\"\"}" \
  --abi "SafeMultisigWallet.abi.json" \
  --sign "msig.keys.json"

tonos-cli deploy \
  SafeMultisigWallet.tvc \
  "{\"owners\":[\"...\"],\"reqConfirms\":...}" \
  --abi SafeMultisigWallet.abi.json \
  --sign contract_keys.json \
  --wc -1

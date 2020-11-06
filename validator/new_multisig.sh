#!/bin/bash

#Nikita Monakhov, [May 9, 2020 at 21:46:32]:
#Можно начать с сид фразы
#genphrase, потом getkeypair, потом genaddr, потом на киньте на адрес токены (один два), потом deploy
#
#А потом потом уже call

tonos-cli genphrase
# Seed phrase: "film slender economy force fox worth soul clock appear transfer wait lawsuit"

tonos-cli getkeypair multisig.keys.json "film slender economy force fox worth soul clock appear transfer wait lawsuit"
cat multisig.keys.json
#{
#  "public": "628f7eaa37116ffd4726a7912c988902df266cf9e43e2c84e2ceb54ce32e74c0",
#  "secret": "835340f9cbd475f7f3e906cf7664656b70f14e7735cc228674cb65c82cc87a2f"
#}

tonos-cli genaddr /data/net.ton.dev/configs/SafeMultisigWallet.tvc /data/net.ton.dev/configs/SafeMultisigWallet.abi.json --setkey multisig.keys.json --wc -1
#Raw address: -1:ad47ea0c469262fe9cbdb07423dfd0e7173dd1dfd940eb2a962335ef85d2b859
#testnet:
#Non-bounceable address (for init): 0f+tR+oMRpJi/py9sHQj39DnFz3R39lA6yqWIzXvhdK4WV9W
#Bounceable address (for later access): kf+tR+oMRpJi/py9sHQj39DnFz3R39lA6yqWIzXvhdK4WQKT
#mainnet:
#Non-bounceable address (for init): Uf+tR+oMRpJi/py9sHQj39DnFz3R39lA6yqWIzXvhdK4WeTc
#Bounceable address (for later access): Ef+tR+oMRpJi/py9sHQj39DnFz3R39lA6yqWIzXvhdK4WbkZ
#Succeeded

"${TON_BUILD_DIR}/crypto/fift" -I "${TON_SRC_DIR}/crypto/fift/lib:${TON_SRC_DIR}/crypto/smartcont" -s wallet.fif validator_2 Uf+tR+oMRpJi/py9sHQj39DnFz3R39lA6yqWIzXvhdK4WeTc 1 1
"${TON_BUILD_DIR}/crypto/fift" -I "${TON_SRC_DIR}/crypto/fift/lib:${TON_SRC_DIR}/crypto/smartcont" -s wallet.fif validator_2 Uf+tR+oMRpJi/py9sHQj39DnFz3R39lA6yqWIzXvhdK4WeTc 1 24997.5


tonos-cli deploy \
  /data/net.ton.dev/configs/SafeMultisigWallet.tvc \
  "{\"owners\":[\"0x628f7eaa37116ffd4726a7912c988902df266cf9e43e2c84e2ceb54ce32e74c0\",\"0x264d1b03503a0a5bc00d64b50232092d6f5ccdc5f137af0f821400464a1a509b\",\"0xf7cdee2eb4dda42a297285b3165515afb56a6073b941f3b8a99b2bd932e8e61c\",\"0x091f94d7b65ce05dff3994ef5598f92a188ec3cb3a34721432e68b873920bf13\",\"0xd272081bfc1d6418b1483559d5e3358e1b5108eb0bfec3b00b570b8605c4041b\"],\"reqConfirms\":4}" \
  --abi /data/net.ton.dev/configs/SafeMultisigWallet.abi.json \
  --sign multisig.keys.json \
  --wc -1

"${UTILS_DIR}/tonos-cli" call "${MSIG_ADDR}" submitTransaction \
  "{\"dest\":\"-1:ad47ea0c469262fe9cbdb07423dfd0e7173dd1dfd940eb2a962335ef85d2b859\",\"value\":\"50000000000000\",\"bounce\":false,\"allBalance\":false,\"payload\":\"\"}" \
  --abi "${CONFIGS_DIR}/SafeMultisigWallet.abi.json" \
  --sign "${KEYS_DIR}/msig.keys.json"

"${UTILS_DIR}/tonos-cli" call "${MSIG_ADDR}" submitTransaction \
  "{\"dest\":\"-1:ad47ea0c469262fe9cbdb07423dfd0e7173dd1dfd940eb2a962335ef85d2b859\",\"value\":\"50000000000000\",\"bounce\":false,\"allBalance\":false,\"payload\":\"\"}" \
  --abi "${CONFIGS_DIR}/SafeMultisigWallet.abi.json" \
  --sign "${KEYS_DIR}/msig.keys.json"

tonos-cli call "-1:8b2f47066d5c00320163064d2af2810637e6dc9a7cc08992a2e236e7ecce289b" \
  confirmTransaction \
  "{\"transactionId\":\"0x5eb723f9f40c7801\"}" \
  --abi "${CONFIGS_DIR}/SafeMultisigWallet.abi.json" \
  --sign "/home/akiselev/ton-keys/key1/msig.keys.json"


1042174221346224701629391720081787574554436045566692307506687838329061435371 24999000000000 24DD98651AD738D1F8BDF58F2AE6F4F5330C53F4F13A179FF7D44891C22E7EB
12793320597111461592807458390059771221641173374306316053960357917551532001366 24999000000000 1C48C34A771898F3460EC1A009FFCA22F43B81608431EDEAB141877440114856
69497531911418355571389143502070955218353123776962244080868677313639778600642 24999000000000 99A635883BF9FA12B9C3FC1BB31115BEBFFC04237654232ACC7A81711997C6C2
73507481921644665564205450764607652770976918062941117120546813399756533864389 24999000000000 A283C2A8A730CAEC94EAA122CA69F5073DB97B4A3804E64FC9E4ADC317DF17C5
77407533297222222636714217849394757893660168830079914237590605052285812245383 24999000000000 AB231C7A27EE54D9A3BFEDE7B8DABD323CA5FC7A28DB09B7F8762313F2F88387
88543357574907674296591025100746979206562164360523574344546753855089257568758 24999000000000 C3C1C3B13191E2B607B26E470656658D521C182A188533244B33BD72C79959F6
102035246701738212792650909306180739549504956425155674957550381851681650718698 24999000000000 E195E72E87AB54B83DE46EB0CF861949ADB0A929060E6B7F9B6B347E48A537EA

999f9508e5f279bd4b1807a674415b4f9855d7bcbbefa29d8226d059ebc4f668


ADNL: mZ+VCOXyeb1LGAemdEFbT5hV17y776KdgibQWevE9mg=
EL: p0kB1+gJBoQo+k6dQSJAF5Pr9693F+H1BmyKFOe/2TE=

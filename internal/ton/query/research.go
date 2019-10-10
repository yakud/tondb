package query

const (
	queryCheckShardsId = `
	SELECT hex(ShardPrefix), ShardPrefix, ShardPfxBits,
	      hex(toUInt64(bitOr(ShardPrefix, bitShiftLeft(1, (63 - toUInt64(ShardPfxBits)))))) as realShard
	FROM blocks 
	WHERE ShardWorkchainId = 0 
	GROUP BY ShardPrefix, ShardPfxBits 
	ORDER BY ShardPrefix;


	SELECT hex(ShardPrefix), ShardPrefix 
	FROM shards_descr 
	WHERE ShardWorkchainId = 0 
	GROUP BY ShardPrefix
	ORDER BY ShardPrefix;
`
)

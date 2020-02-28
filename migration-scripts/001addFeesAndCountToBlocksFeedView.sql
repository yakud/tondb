CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_BlocksFeedNew
ENGINE = MergeTree()
PARTITION BY toStartOfYear(Time)
ORDER BY (Time, WorkchainId, Shard, SeqNo)
SETTINGS index_granularity=128,index_granularity_bytes=0
POPULATE
AS
SELECT
    WorkchainId,
    Shard,
    SeqNo,
    Time
FROM blocks;

RENAME TABLE _view_feed_BlocksFeed TO _view_feed_BlocksFeedOld, _view_feed_BlocksFeedNew TO _view_feed_BlocksFeed;

DROP TABLE _view_feed_BlocksFeedOld;

-- This table will be created anyways at new version of ton-api start so it can be not executed
CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_TransactionFeesFeed
ENGINE = SummingMergeTree()
PARTITION BY toStartOfYear(Time)
ORDER BY (WorkchainId, Shard, SeqNo)
SETTINGS index_granularity=128,index_granularity_bytes=0
POPULATE
AS
SELECT
    Time,
    TotalFeesNanograms,
    WorkchainId,
    Shard,
    SeqNo,
    count() AS Count,
    sumArray(Messages.ValueNanograms) AS ValueNanograms,
    sumArray(Messages.IhrFeeNanograms) AS IhrFeeNanograms,
    sumArray(Messages.ImportFeeNanograms) AS ImportFeeNanograms,
    sumArray(Messages.FwdFeeNanograms) AS FwdFeeNanograms
FROM transactions
GROUP BY Time, TotalFeesNanograms, WorkchainId, Shard, SeqNo;
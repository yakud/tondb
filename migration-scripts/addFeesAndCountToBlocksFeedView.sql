CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_BlocksFeedNew
    ENGINE = SummingMergeTree()
        PARTITION BY toStartOfYear(Time)
        ORDER BY (Time, WorkchainId, Shard, SeqNo)
        SETTINGS index_granularity=128,index_granularity_bytes=0
    POPULATE
AS
SELECT
    WorkchainId,
    Shard,
    SeqNo,
    Time,
    TotalFeesNanograms,
    count,
    ValueNanograms,
    IhrFeeNanograms,
    ImportFeeNanograms,
    FwdFeeNanograms
FROM (
     SELECT
         WorkchainId,
         Shard,
         SeqNo,
         Time
     FROM blocks
     ) ANY LEFT JOIN (
SELECT
    TotalFeesNanograms,
    WorkchainId,
    Shard,
    SeqNo,
    count() AS count,
    sumArray(Messages.ValueNanograms) AS ValueNanograms,
    sumArray(Messages.IhrFeeNanograms) AS IhrFeeNanograms,
    sumArray(Messages.ImportFeeNanograms) AS ImportFeeNanograms,
    sumArray(Messages.FwdFeeNanograms) AS FwdFeeNanograms
FROM transactions GROUP BY TotalFeesNanograms, WorkchainId, Shard, SeqNo
) USING (WorkchainId, Shard, SeqNo);

RENAME TABLE _view_feed_BlocksFeed TO _view_feed_BlocksFeedOld, _view_feed_BlocksFeedNew TO _view_feed_BlocksFeed;

DROP TABLE _view_feed_BlocksFeedOld;
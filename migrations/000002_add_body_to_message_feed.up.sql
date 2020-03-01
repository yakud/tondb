CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_MessagesFeedGlobalNew
ENGINE = MergeTree()
PARTITION BY toYYYYMM(Time)
ORDER BY (Time, WorkchainId, Lt, MessageLt)
SETTINGS index_granularity=128,index_granularity_bytes=0
POPULATE
AS
SELECT
    WorkchainId,
    Shard,
    SeqNo,
    Lt,
    Time,
    Hash AS TrxHash,
    Messages.CreatedLt as MessageLt,
    Messages.Direction as Direction,
    Messages.SrcWorkchainId AS SrcWorkchainId,
    Messages.SrcAddr AS Src,
    Messages.DestWorkchainId AS DestWorkchainId,
    Messages.DestAddr AS Dest,
    Messages.ValueNanograms as ValueNanograms,
    Messages.FwdFeeNanograms + Messages.IhrFeeNanograms + Messages.ImportFeeNanograms as TotalFeeNanograms,
    Messages.Bounce as Bounce,
    Messages.BodyType as BodyType,
    Messages.BodyValue as BodyValue
FROM transactions
ARRAY JOIN Messages
WHERE Type = 'trans_ord' AND Messages.Type = 'int_msg_info';

RENAME TABLE _view_feed_MessagesFeedGlobal TO _view_feed_MessagesFeedGlobalOld, _view_feed_MessagesFeedGlobalNew TO _view_feed_MessagesFeedGlobal;
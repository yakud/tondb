CREATE TABLE IF NOT EXISTS shards_descr (
    MasterShard      UInt64,
    MasterSeqNo      UInt64,
    ShardWorkchainId Int32,
    Shard            UInt64,
    ShardSeqNo       UInt64
)
ENGINE MergeTree
ORDER BY (MasterSeqNo, Shard, ShardSeqNo);

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_state_AccountState
ENGINE = ReplacingMergeTree(SeqNo)
ORDER BY (WorkchainId, Addr)
SETTINGS index_granularity = 64
POPULATE
AS
SELECT
    *
FROM account_state;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_AccountTransactions
ENGINE = MergeTree()
PARTITION BY toYYYYMM(Time)
ORDER BY (WorkchainId, AccountAddr, Lt, Time)
SETTINGS index_granularity=128,index_granularity_bytes=0
POPULATE
AS
SELECT
    WorkchainId,
    Shard,
    SeqNo,
    AccountAddr,
    Lt,
    Time,
    Type,
    Messages.Type as MessageType,
    Messages.CreatedLt as MessageLt,
    Messages.Direction as Direction,
    Messages.SrcWorkchainId AS SrcWorkchainId,
    Messages.SrcAddr AS Src,
    Messages.DestWorkchainId AS DestWorkchainId,
    Messages.DestAddr AS Dest,
    Messages.ValueNanograms as ValueNanograms,
    Messages.FwdFeeNanograms as FwdFeeNanograms,
    Messages.IhrFeeNanograms as IhrFeeNanograms,
    Messages.ImportFeeNanograms as ImportFeeNanograms,
    Messages.Bounce as Bounce,
    Messages.Bounced as Bounced
FROM transactions
ARRAY JOIN Messages;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_BlocksFeed
ENGINE = MergeTree()
PARTITION BY toStartOfYear(Time)
ORDER BY (Time, StartLt, Shard, WorkchainId)
SETTINGS index_granularity=64,index_granularity_bytes=0
POPULATE
AS
SELECT
    WorkchainId,
    Shard,
    SeqNo,
    Time,
    StartLt,
    RootHash
FROM blocks;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_TransactionFeesFeed
ENGINE = SummingMergeTree()
PARTITION BY toStartOfYear(Time)
ORDER BY (Time, WorkchainId, Shard, SeqNo)
SETTINGS index_granularity=128,index_granularity_bytes=0
POPULATE
AS
SELECT
    Time,
    TotalFeesNanograms,
    WorkchainId,
    Shard,
    SeqNo,
    count() AS TrxCount,
    sumArray(Messages.ValueNanograms) AS ValueNanograms,
    sumArray(Messages.IhrFeeNanograms) AS IhrFeeNanograms,
    sumArray(Messages.ImportFeeNanograms) AS ImportFeeNanograms,
    sumArray(Messages.FwdFeeNanograms) AS FwdFeeNanograms
FROM transactions
GROUP BY Time, TotalFeesNanograms, WorkchainId, Shard, SeqNo;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_MessagesFeedGlobal
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
    Messages.Bounce as Bounce
FROM transactions
ARRAY JOIN Messages
WHERE Type = 'trans_ord' AND Messages.Type = 'int_msg_info';

CREATE MATERIALIZED VIEW IF NOT EXISTS _ts_BlocksByWorkchain
ENGINE = SummingMergeTree()
PARTITION BY tuple()
ORDER BY (Time, WorkchainId)
POPULATE
AS
SELECT
    toStartOfInterval(Time, INTERVAL 5 SECOND) as Time,
    WorkchainId,
    count() as Blocks
FROM blocks
GROUP BY Time, WorkchainId;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_ts_MessagesByType
ENGINE = SummingMergeTree()
PARTITION BY tuple()
ORDER BY (Time, WorkchainId, Type, MsgType)
POPULATE
AS
SELECT
    toStartOfInterval(Time, INTERVAL 5 MINUTE) as Time,
    WorkchainId,
    Type,
    Messages.Type as MsgType,
    count() as MessagesCount
FROM transactions
ARRAY JOIN Messages
GROUP BY Time, WorkchainId, Type, MsgType;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_ts_VolumeByGrams
ENGINE = SummingMergeTree()
PARTITION BY tuple()
ORDER BY (Time, WorkchainId)
POPULATE
AS
SELECT
    toStartOfInterval(Time, INTERVAL 10 MINUTE) as Time,
    WorkchainId,
    sum(Messages.ValueNanograms) as VolumeNanograms
FROM transactions
ARRAY JOIN Messages
GROUP BY Time, WorkchainId;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_ts_MessagesOrdCount
ENGINE = SummingMergeTree()
PARTITION BY tuple()
ORDER BY (Time, WorkchainId)
POPULATE
AS
SELECT
    toStartOfInterval(Time, INTERVAL 10 MINUTE) as Time,
    WorkchainId,
    count() as MessagesCount
FROM transactions
ARRAY JOIN Messages
WHERE
    Type = 'trans_ord' AND
    Messages.ValueNanograms > 0
GROUP BY Time, WorkchainId;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_stats_AddrMessagesCountTop
ENGINE = SummingMergeTree()
PARTITION BY tuple()
ORDER BY (Direction, Addr, WorkchainId)
POPULATE
AS
SELECT
    Messages.Direction as Direction,
    if(Direction = 'in', Messages.DestAddr, Messages.SrcAddr) AS Addr,
    WorkchainId,
    count() as Count
FROM transactions
ARRAY JOIN Messages
WHERE Type = 'trans_ord' AND Messages.Type = 'int_msg_info'
GROUP BY Direction, Addr, WorkchainId;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_MessagesPerSecond
ENGINE = MergeTree()
PARTITION BY toStartOfYear(Time)
ORDER BY (Time, WorkchainId)
POPULATE AS
SELECT
    WorkchainId,
    Time,
    count() AS TrxCount,
    sum(length(Messages.Direction)) AS MsgCount
FROM transactions
GROUP BY Time, WorkchainId;

CREATE MATERIALIZED VIEW IF NOT EXISTS _view_feed_TotalTransactionsAndMessages
ENGINE = SummingMergeTree()
PARTITION BY tuple()
ORDER BY (WorkchainId)
POPULATE
AS
SELECT
    WorkchainId,
    count() as TotalTransactions,
    sum(length(Messages.Direction)) AS TotalMessages
FROM transactions
GROUP BY WorkchainId;




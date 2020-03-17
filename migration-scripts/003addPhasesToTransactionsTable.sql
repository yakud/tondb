CREATE TABLE IF NOT EXISTS transactionsNew (
    WorkchainId           Int32,
    Shard                 UInt64,
    SeqNo                 UInt64,

    Hash                  FixedString(64),
    Type                  LowCardinality(String),
    Lt                    UInt64,
    Time                  DateTime, -- field Now
    TotalFeesNanograms    UInt64,
    TotalFeesNanogramsLen UInt8,
    AccountAddr           FixedString(64),
    OrigStatus            LowCardinality(String),
    EndStatus             LowCardinality(String),
    PrevTransLt   		  UInt64,
    PrevTransHash 		  FixedString(64),
    StateUpdateNewHash    FixedString(64),
    StateUpdateOldHash    FixedString(64),
    IsTock                UInt8,

    ActionPhaseExists             UInt8,
    ActionPhaseSuccess            UInt8,
    ActionPhaseValid              UInt8,
    ActionPhaseNoFunds            UInt8,
    ActionPhaseCodeChanged        UInt8,
    ActionPhaseActionListInvalid  UInt8,
    ActionPhaseAccDeleteReq       UInt8,
    ActionPhaseAccStatusChange    LowCardinality(String),
    ActionPhaseTotalFwdFees       UInt64,
    ActionPhaseTotalActionFees    UInt64,
    ActionPhaseResultCode         Int32,
    ActionPhaseResultArg          Int32,
    ActionPhaseTotActions         UInt32,
    ActionPhaseSpecActions        UInt32,
    ActionPhaseSkippedActions     UInt32,
    ActionPhaseMsgsCreated        UInt32,
    ActionPhaseRemainingBalance   UInt64,
    ActionPhaseReservedBalance    UInt64,
    ActionPhaseEndLt              UInt64,
    ActionPhaseTotMsgBits         UInt64,
    ActionPhaseTotMsgCells        UInt64,

    ComputePhaseExists           UInt8,
    ComputePhaseSkipped          UInt8,
    ComputePhaseSkippedReason    LowCardinality(String),
    ComputePhaseAccountActivated UInt8,
    ComputePhaseSuccess          UInt8,
    ComputePhaseMsgStateUsed     UInt8,
    ComputePhaseOutOfGas         UInt8,
    ComputePhaseAccepted         UInt8,
    ComputePhaseExitArg          Int32,
    ComputePhaseExitCode         Int32,
    ComputePhaseMode             Int8,
    ComputePhaseVmSteps          UInt32,
    ComputePhaseGasUsed          UInt64,
    ComputePhaseGasMax           UInt64,
    ComputePhaseGasCredit        UInt64,
    ComputePhaseGasLimit         UInt64,
    ComputePhaseGasFees          UInt64,

    StoragePhaseExists        UInt8,
    StoragePhaseStatus        LowCardinality(String),
    StoragePhaseFeesCollected UInt64,
    StoragePhaseFeesDue       UInt64,

    CreditPhaseExists           UInt8,
    CreditPhaseDueFeesCollected UInt64,
    CreditPhaseCreditNanograms  UInt64,

    Messages Nested
    (
        Direction             LowCardinality(String),
        Type                  LowCardinality(String),
        Init                  LowCardinality(String),
        Bounce                UInt8,
        Bounced               UInt8,
        CreatedAt             UInt64,
        CreatedLt             UInt64,
        ValueNanograms        UInt64,
        ValueNanogramsLen     UInt8,
        FwdFeeNanograms       UInt64,
        FwdFeeNanogramsLen    UInt8,
        IhrDisabled           UInt8,
        IhrFeeNanograms       UInt64,
        IhrFeeNanogramsLen    UInt8,
        ImportFeeNanograms    UInt64,
        ImportFeeNanogramsLen UInt8,
        DestIsEmpty           UInt8,
        DestWorkchainId       Int32,
        DestAddr              FixedString(64),
        DestAnycast           LowCardinality(String),
        SrcIsEmpty            UInt8,
        SrcWorkchainId        Int32,
        SrcAddr               FixedString(64),
        SrcAnycast            LowCardinality(String),
        BodyType              LowCardinality(String),
        BodyValue             String
    )
) ENGINE MergeTree
PARTITION BY toYYYYMM(Time)
ORDER BY (WorkchainId, Shard, SeqNo, Lt);

RENAME TABLE transactions TO transactionsOld, transactionsNew TO transactions;

DROP TABLE transactionsOld;
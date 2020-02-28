ALTER TABLE blocks ADD COLUMN IF NOT EXISTS ValueFlowFromPrevBlk UInt64 AFTER BeforeSplit;

ALTER TABLE blocks ADD COLUMN IF NOT EXISTS ValueFlowToNextBlk UInt64 AFTER ValueFlowFromPrevBlk;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS ValueFlowImported UInt64 AFTER ValueFlowToNextBlk;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS ValueFlowExported UInt64 AFTER ValueFlowImported;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS ValueFlowFeesCollected UInt64 AFTER ValueFlowExported;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS ValueFlowFeesImported UInt64 AFTER ValueFlowFeesCollected;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS ValueFlowRecovered UInt64 AFTER ValueFlowFeesImported;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS ValueFlowCreated UInt64 AFTER ValueFlowRecovered;
ALTER TABLE blocks ADD COLUMN IF NOT EXISTS ValueFlowMinted UInt64 AFTER ValueFlowCreated;

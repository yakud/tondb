package writer

//type OnBulkBlockWriteHandler func(blockNum uint32) error

// Collect action traces and states for exactly one block.
// Needed for full block atomic writes.
//type BulkBlock struct {
//	blockNum uint32
//
//	blocks []*ton.Block
//
//	onWriteHandler OnBulkBlockWriteHandler
//}
//
//func (t *BulkBlock) BlockNum() uint32 {
//	return t.blockNum
//}
//
//func (t *BulkBlock) SetBlockNum(blockNum uint32) {
//	t.blockNum = blockNum
//}
//
//func (t *BulkBlock) AddBlocks(block ...*ton.Block) {
//	t.blocks = append(t.blocks, block...)
//}
//
//func (t *BulkBlock) Blocks() []*ton.Block {
//	return t.blocks
//}
//
//func (t *BulkBlock) LengthBlocks() int {
//	return len(t.blocks)
//}
//
//func (t *BulkBlock) LengthTransactions() int {
//	total := 0
//	for _, b := range t.blocks {
//		total += len(b.Transactions)
//	}
//	return total
//}
//
//func (t *BulkBlock) LengthTotal() int {
//	return t.LengthBlocks() + t.LengthTransactions()
//}
//
//func (t *BulkBlock) SetOnWrite(onWriteHandler OnBulkBlockWriteHandler) {
//	t.onWriteHandler = onWriteHandler
//}
//
//func (t *BulkBlock) FireOnWrite() error {
//	if t.onWriteHandler != nil {
//		return t.onWriteHandler(t.blockNum)
//	}
//
//	return nil
//}
//
//func NewBulkBlock() *BulkBlock {
//	return &BulkBlock{
//		blocks: make([]*ton.Block, 0),
//	}
//}

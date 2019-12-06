package tlb_pretty

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

// Convert AST to ton.block
type AstTonConverter struct {
}

func (c *AstTonConverter) ConvertToBlock(node *AstNode) (*ton.Block, error) {
	block := &ton.Block{
		Transactions: make([]*ton.Transaction, 0),
	}

	// Block info
	var err error
	block.Info, err = c.extractBlockInfo(node)
	if err != nil {
		return nil, err
	}

	// Block header
	blockHeader, err := c.extractBlockHeader(node)
	if err != nil {
		return nil, err
	}
	block.Info.BlockHeader = *blockHeader

	// Transactions
	if err = c.extractTransactions(node, &block.Transactions); err != nil {
		return nil, err
	}

	for _, tr := range block.Transactions {
		tr.WorkchainId = block.Info.WorkchainId
		tr.Shard = block.Info.Shard
		tr.SeqNo = block.Info.SeqNo
	}

	if err := c.extractTransactionsHash(node, &block.Transactions); err != nil {
		return nil, err
	}

	sort.SliceStable(block.Transactions, func(i, j int) bool {
		return block.Transactions[i].Lt < block.Transactions[j].Lt
	})

	// shard_hashes only for workchain -1
	if block.Info.WorkchainId == -1 {
		shardsDescr, err := c.extractShardsDescr(node)
		if err != nil {
			return nil, err
		}
		for _, descr := range shardsDescr {
			descr.ShardWorkchainId = 0 // todo: can't find workchain field in data
			descr.MasterShard = block.Info.Shard
			descr.MasterSeqNo = block.Info.SeqNo
		}
		block.ShardDescr = shardsDescr
	}

	return block, nil
}

func (c *AstTonConverter) extractShardsDescr(node *AstNode) ([]*ton.ShardDescr, error) {
	customNode, err := node.GetNode("extra", "custom", "value")
	if err != nil {
		return nil, nil
	}

	if !customNode.IsType("masterchain_block_extra") {
		fmt.Println("custom.value is not masterchain_block_extra type")
		return nil, nil
	}

	shardsDescrRoot, err := customNode.GetNode("shard_hashes", "value_0")
	if err != nil {
		return nil, err
	}

	if !shardsDescrRoot.IsType("hme_root") {
		return nil, errors.New("shards descr root node is not hme_root type")
	}

	shardDescrs := make([]*ton.ShardDescr, 0)

	err = shardsDescrRoot.EachNode(func(i int, el *AstNode) error {
		leafNode, err := el.GetNode("leaf")
		if err != nil {
			return err
		}

		shardDescr := &ton.ShardDescr{
			ShardWorkchainId: 0,
		}

		shardDescr.Shard, err = leafNode.GetUint64("next_validator_shard")
		if err != nil {
			return err
		}

		shardDescr.ShardSeqNo, err = leafNode.GetUint64("seq_no")
		if err != nil {
			return err
		}

		shardDescrs = append(shardDescrs, shardDescr)

		return nil
	}, "leafs")
	if err != nil {
		return nil, err
	}

	return shardDescrs, nil
}

func (c *AstTonConverter) extractBlockHeader(node *AstNode) (*ton.BlockHeader, error) {
	var err error
	header := &ton.BlockHeader{}

	nodeHeader, err := node.GetNode("header")
	if err != nil {
		return header, nil
	}

	if header.RootHash, err = nodeHeader.GetString("root_hash"); err != nil {
		return nil, err
	}

	if header.FileHash, err = nodeHeader.GetString("file_hash"); err != nil {
		return nil, err
	}

	return header, nil
}

func (c *AstTonConverter) extractBlockInfo(node *AstNode) (*ton.BlockInfo, error) {
	var err error
	info := &ton.BlockInfo{}

	nodeInfo, err := node.GetNode("info")
	if err != nil {
		return nil, err
	}

	nodeShard, err := nodeInfo.GetNode("shard")
	if err != nil {
		return nil, err
	}

	// Shard
	if info.WorkchainId, err = nodeShard.GetInt32("workchain_id"); err != nil {
		return nil, err
	}

	var shardPrefix uint64
	var shardPfxBits uint8
	if shardPrefix, err = nodeShard.GetUint64("shard_prefix"); err != nil {
		return nil, err
	}
	if shardPfxBits, err = nodeShard.GetUint8("shard_pfx_bits"); err != nil {
		return nil, err
	}
	info.Shard = c.ConvertShardPrefixToShard(shardPrefix, shardPfxBits)

	if info.SeqNo, err = nodeInfo.GetUint64("seq_no"); err != nil {
		return nil, err
	}

	// Flags, some data
	if info.MinRefMcSeqno, err = nodeInfo.GetUint32("min_ref_mc_seqno"); err != nil {
		return nil, err
	}
	if info.PrevKeyBlockSeqno, err = nodeInfo.GetUint32("prev_key_block_seqno"); err != nil {
		return nil, err
	}
	if info.GenCatchainSeqno, err = nodeInfo.GetUint32("gen_catchain_seqno"); err != nil {
		return nil, err
	}
	if info.GenUtime, err = nodeInfo.GetUint32("gen_utime"); err != nil {
		return nil, err
	}
	if info.StartLt, err = nodeInfo.GetUint64("start_lt"); err != nil {
		return nil, err
	}
	if info.EndLt, err = nodeInfo.GetUint64("end_lt"); err != nil {
		return nil, err
	}
	if info.Version, err = nodeInfo.GetUint32("version"); err != nil {
		return nil, err
	}
	if info.Flags, err = nodeInfo.GetUint8("flags"); err != nil {
		return nil, err
	}
	if info.KeyBlock, err = nodeInfo.GetBool("key_block"); err != nil {
		return nil, err
	}
	if info.NotMaster, err = nodeInfo.GetBool("not_master"); err != nil {
		return nil, err
	}
	if info.WantMerge, err = nodeInfo.GetBool("want_merge"); err != nil {
		return nil, err
	}
	if info.WantSplit, err = nodeInfo.GetBool("want_split"); err != nil {
		return nil, err
	}
	if info.AfterMerge, err = nodeInfo.GetBool("after_merge"); err != nil {
		return nil, err
	}
	if info.AfterSplit, err = nodeInfo.GetBool("after_split"); err != nil {
		return nil, err
	}
	if info.BeforeSplit, err = nodeInfo.GetBool("before_split"); err != nil {
		return nil, err
	}

	// Prev ref
	prevRefNode, err := nodeInfo.GetNode("prev_ref", "prev")
	if err == nil {
		info.Prev1Ref = &ton.BlockRef{}
		if info.Prev1Ref.EndLt, err = prevRefNode.GetUint64("end_lt"); err != nil {
			return nil, err
		}
		if info.Prev1Ref.SeqNo, err = prevRefNode.GetUint64("seq_no"); err != nil {
			return nil, err
		}
		if info.Prev1Ref.FileHash, err = prevRefNode.GetString("file_hash"); err != nil {
			return nil, err
		}
		if info.Prev1Ref.RootHash, err = prevRefNode.GetString("root_hash"); err != nil {
			return nil, err
		}
	} else {
		prevRef1Node, err := nodeInfo.GetNode("prev_ref", "prev1")
		if err != nil {
			return nil, err
		}
		info.Prev1Ref = &ton.BlockRef{}
		if info.Prev1Ref.EndLt, err = prevRef1Node.GetUint64("end_lt"); err != nil {
			return nil, err
		}
		if info.Prev1Ref.SeqNo, err = prevRef1Node.GetUint64("seq_no"); err != nil {
			return nil, err
		}
		if info.Prev1Ref.FileHash, err = prevRef1Node.GetString("file_hash"); err != nil {
			return nil, err
		}
		if info.Prev1Ref.RootHash, err = prevRef1Node.GetString("root_hash"); err != nil {
			return nil, err
		}

		prevRef2Node, err := nodeInfo.GetNode("prev_ref", "prev2")
		if err != nil {
			return nil, err
		}
		info.Prev2Ref = &ton.BlockRef{}
		if info.Prev2Ref.EndLt, err = prevRef2Node.GetUint64("end_lt"); err != nil {
			return nil, err
		}
		if info.Prev2Ref.SeqNo, err = prevRef2Node.GetUint64("seq_no"); err != nil {
			return nil, err
		}
		if info.Prev2Ref.FileHash, err = prevRef2Node.GetString("file_hash"); err != nil {
			return nil, err
		}
		if info.Prev2Ref.RootHash, err = prevRef2Node.GetString("root_hash"); err != nil {
			return nil, err
		}
	}

	// Master ref
	masterRefNode, err := nodeInfo.GetNode("master_ref", "master")
	if err == nil {
		info.MasterRef = &ton.BlockRef{}
		if info.MasterRef.EndLt, err = masterRefNode.GetUint64("end_lt"); err != nil {
			return nil, err
		}
		if info.MasterRef.SeqNo, err = masterRefNode.GetUint64("seq_no"); err != nil {
			return nil, err
		}
		if info.MasterRef.FileHash, err = masterRefNode.GetString("file_hash"); err != nil {
			return nil, err
		}
		if info.MasterRef.RootHash, err = masterRefNode.GetString("root_hash"); err != nil {
			return nil, err
		}
	}

	return info, nil
}

func (c *AstTonConverter) extractTransactions(node *AstNode, transactions *[]*ton.Transaction) error {
	accountBlocksRoot, err := node.GetNode("extra", "account_blocks", "value_0")
	if err != nil {
		return err
	}

	if accountBlocksRoot.IsType("ahme_empty") {
		return nil
	}

	return accountBlocksRoot.EachNode(func(i int, el *AstNode) error {
		leafValueNode, err := el.GetNode("value")
		if err != nil {
			return err
		}

		if !leafValueNode.IsType("acc_trans") {
			return errors.New("leafs.value is not acc_trans type")
		}

		transactionsNode, err := leafValueNode.GetNode("transactions", "node")
		if err != nil {
			return err
		}

		return c.extractTransaction(transactionsNode, transactions)
	}, "leafs")
}

func (c *AstTonConverter) extractTransaction(node *AstNode, transactions *[]*ton.Transaction) error {
	leafValueType, err := node.Type()
	if err != nil {
		return err
	}

	if leafValueType == "ahmn_fork" {
		if leftNode, err := node.GetNode("left", "node"); err == nil {
			if err := c.extractTransaction(leftNode, transactions); err != nil {
				return err
			}
		}
		if rightNode, err := node.GetNode("right", "node"); err == nil {
			if err := c.extractTransaction(rightNode, transactions); err != nil {
				return err
			}
		}
	} else if leafValueType == "ahmn_leaf" {
		transactionNode, err := node.GetNode("value")
		if err != nil {
			return err
		}

		tr := &ton.Transaction{}

		if tr.Lt, err = transactionNode.GetUint64("lt"); err != nil {
			return err
		}
		if tr.Type, err = transactionNode.GetString("description", "@type"); err != nil {
			return err
		}
		if tr.Now, err = transactionNode.GetUint64("now"); err != nil {
			return err
		}
		if tr.AccountAddr, err = transactionNode.GetString("account_addr"); err != nil {
			return err
		}
		if tr.OrigStatus, err = transactionNode.GetString("orig_status"); err != nil {
			return err
		}
		if tr.EndStatus, err = transactionNode.GetString("end_status"); err != nil {
			return err
		}
		if tr.PrevTransLt, err = transactionNode.GetUint64("prev_trans_lt"); err != nil {
			return err
		}
		if tr.PrevTransHash, err = transactionNode.GetString("prev_trans_hash"); err != nil {
			return err
		}
		if tr.StateUpdateNewHash, err = transactionNode.GetString("state_update", "new_hash"); err != nil {
			return err
		}
		if tr.StateUpdateOldHash, err = transactionNode.GetString("state_update", "old_hash"); err != nil {
			return err
		}
		if tr.TotalFeesNanograms, err = transactionNode.GetUint64("total_fees", "grams", "amount", "value"); err != nil {
			return err
		}
		if tr.TotalFeesNanogramsLen, err = transactionNode.GetUint8("total_fees", "grams", "amount", "len"); err != nil {
			return err
		}

		//switch tr.Type {
		//case "trans_ord":
		//	if tr.Type, err = transactionNode.GetString("description", "action"); err != nil {
		//		return err
		//	}
		//case "trans_tick_tock":
		//
		//}

		// In message extract
		if inMsgNode, err := transactionNode.GetNode("value_0", "in_msg", "value"); err != nil {
			//if inMsgStr, err := transactionNode.GetString("value_0", "in_msg"); err != nil {
			//	return err
			//} else if inMsgStr != "hme_empty" && inMsgStr != "nothing" {
			//	return errors.New("undefined in_msg type:" + inMsgStr)
			//}
		} else {
			if tr.InMsg, err = c.extractMessage(inMsgNode); err != nil {
				//return fmt.Errorf("in_msg err: %+v", err)
			}
		}

		// Out message extract
		if outMsgNode, err := transactionNode.GetNode("value_0", "out_msgs"); err != nil {
			//if outMsgsStr, err := transactionNode.GetString("value_0", "out_msgs"); err != nil {
			//	return err
			//} else if outMsgsStr != "hme_empty" && outMsgsStr != "nothing" {
			//	return errors.New("undefined out_msgs type")
			//}
		} else {
			err = outMsgNode.EachNode(func(i int, el *AstNode) error {
				outMsgValueNode, err := el.GetNode("value")
				if err != nil {
					return err
				}

				msg, err := c.extractMessage(outMsgValueNode)
				if err != nil {
					return err
				}

				tr.OutMsgs = append(tr.OutMsgs, msg)
				return nil
			}, "leafs")

			if err != nil {
				//return fmt.Errorf("out_msgs err: %+v", err)
			}
		}

		*transactions = append(*transactions, tr)
	} else {
		return errors.New("undefined node type: " + leafValueType)
	}

	return nil
}

func (c *AstTonConverter) extractTransactionsHash(node *AstNode, transaction *[]*ton.Transaction) error {
	transactionsHashNode, err := node.GetNode("transactions_hash")
	if err != nil && len(*transaction) > 0 {
		return fmt.Errorf("not found transactions hashes: %w", err)
	}
	if transactionsHashNode == nil {
		if len(*transaction) > 0 {
			return fmt.Errorf("empty transactions_hash")
		}
		return nil
	}

	for _, tr := range *transaction {
		accLtHashNode, err := transactionsHashNode.GetNode(strings.TrimLeft(tr.AccountAddr, "x"))
		if err != nil {
			return fmt.Errorf("not found lthash node for account %s: %w", tr.AccountAddr, err)
		}

		trHash, err := accLtHashNode.GetString(strconv.FormatUint(tr.Lt, 10))
		if err != nil {
			return fmt.Errorf("not found tr hash for account: %s lt: %d: %w", tr.AccountAddr, tr.Lt, err)
		}

		tr.Hash = trHash
	}

	return nil
}

func (c *AstTonConverter) extractMessage(node *AstNode) (*ton.TransactionMessage, error) {
	if !node.IsType("message") {
		tp, _ := node.Type()
		return nil, fmt.Errorf("is not message type: %s", tp)
	}

	msg := &ton.TransactionMessage{
		IhrDisabled: false,
	}
	var err error

	msgInfoNode, err := node.GetNode("info")
	if err != nil {
		return nil, err
	}

	if msg.Type, err = msgInfoNode.Type(); err != nil {
		return nil, err
	}

	switch msg.Type {
	case "ext_out_msg_info":

	case "ext_in_msg_info":
		if msg.ImportFeeNanograms, err = msgInfoNode.GetUint64("import_fee", "amount", "value"); err != nil {
			return nil, err
		}

	case "int_msg_info":
		if msg.Bounce, err = msgInfoNode.GetBool("bounce"); err != nil {
			return nil, err
		}
		if msg.Bounced, err = msgInfoNode.GetBool("bounced"); err != nil {
			return nil, err
		}
		if msg.CreatedAt, err = msgInfoNode.GetUint64("created_at"); err != nil {
			return nil, err
		}
		if msg.CreatedLt, err = msgInfoNode.GetUint64("created_lt"); err != nil {
			return nil, err
		}
		if msg.ValueNanograms, err = msgInfoNode.GetUint64("value", "grams", "amount", "value"); err != nil {
			return nil, err
		}
		if msg.ValueNanogramsLen, err = msgInfoNode.GetUint8("value", "grams", "amount", "len"); err != nil {
			return nil, err
		}
		if msg.FwdFeeNanograms, err = msgInfoNode.GetUint64("fwd_fee", "amount", "value"); err != nil {
			return nil, err
		}
		if msg.FwdFeeNanogramsLen, err = msgInfoNode.GetUint8("fwd_fee", "amount", "len"); err != nil {
			return nil, err
		}
		if msg.IhrFeeNanograms, err = msgInfoNode.GetUint64("ihr_fee", "amount", "value"); err != nil {
			return nil, err
		}
		if msg.IhrFeeNanogramsLen, err = msgInfoNode.GetUint8("ihr_fee", "amount", "len"); err != nil {
			return nil, err
		}
		if msg.IhrDisabled, err = msgInfoNode.GetBool("ihr_disabled"); err != nil {
			return nil, err
		}

	default:
		//return nil, fmt.Errorf("undefined transaction message type: %s", msg.Type)
	}

	if init, err := msgInfoNode.GetString("init"); err == nil {
		msg.Init = init
	}

	if msg.BodyValue, err = node.GetString("body", "value", "value"); err != nil {
		return nil, err
	}
	if msg.BodyType, err = node.GetString("body", "value", "@type"); err != nil {
		return nil, err
	}

	if dest, err := msgInfoNode.GetNode("dest"); err != nil {
		msg.Dest.IsEmpty = true
	} else {
		if msg.Dest, err = c.extractAddrStd(dest); err != nil {
			//return nil, err
		}
	}

	if src, err := msgInfoNode.GetNode("src"); err != nil {
		msg.Src.IsEmpty = true
	} else {
		if msg.Src, err = c.extractAddrStd(src); err != nil {
			//return nil, err
		}
	}

	return msg, nil
}

func (c *AstTonConverter) extractAddrStd(node *AstNode) (addr ton.AddrStd, err error) {
	if !node.IsType("addr_std") {
		return addr, errors.New("node is not addr_std type")
	}

	if addr.Addr, err = node.GetString("address"); err != nil {
		return
	}
	if addr.Anycast, err = node.GetString("anycast"); err != nil {
		return
	}
	if addr.WorkchainId, err = node.GetInt32("workchain_id"); err != nil {
		return
	}

	return
}

func (c *AstTonConverter) ConvertShardPrefixToShard(shardPrefix uint64, shardPfxBits uint8) uint64 {
	return shardPrefix | (1 << (63 - shardPfxBits))
}

func NewAstTonConverter() *AstTonConverter {
	return &AstTonConverter{}
}

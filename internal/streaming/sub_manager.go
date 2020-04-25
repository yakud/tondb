package streaming

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"gitlab.flora.loc/mills/tondb/internal/utils"
	"strings"
	"sync"
	"time"
)

type SubManager struct {
	subs map[Filter][]*SubHandler
	rw   sync.RWMutex
}

func (m *SubManager) HandleBlock(block *ton.Block) error {
	indexer := NewTransactionsAndMessagesIndexer()

	blockJson, err := json.Marshal(blockToFeedBlock(block))
	if err != nil {
		return err
	}

	transactions := make([]*feed.TransactionInFeed, 0, len(block.Transactions))
	messages := make([]*feed.MessageInFeed, 0, 2*len(block.Transactions))
	for _, trx := range block.Transactions {
		var trxFeed *feed.TransactionInFeed
		if trxFeed, err = trxToFeedTrx(trx); err != nil {
			return err
		}
		transactions = append(transactions, trxFeed)
		trxJson, err := json.Marshal(trxFeed)
		if err != nil {
			return err
		}

		indexer.AddTransaction(trxFeed, trxJson)

		if trx.InMsg != nil {
			msg, err := messageToFeedMessage(block, trx.InMsg, "in", trx.Lt)
			if err != nil {
				return err
			}

			msgJson, err := json.Marshal(msg)
			if err != nil {
				return err
			}

			indexer.AddMessage(msg, msgJson)
			messages = append(messages, msg)
		}
		if trx.OutMsgs != nil {
			for _, msg := range trx.OutMsgs {
				msgFeed, err := messageToFeedMessage(block, msg, "out", trx.Lt)
				if err != nil {
					return err
				}

				msgJson, err := json.Marshal(msgFeed)
				if err != nil {
					return err
				}

				indexer.AddMessage(msgFeed, msgJson)

				messages = append(messages, msgFeed)
			}
		}
	}

	transactionsJson, err := json.Marshal(transactions)
	if err != nil {
		return err
	}

	messagesJson, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	m.rw.RLock()
	for filter, subs := range m.subs {
		if filter.MatchWorkchainAndShard(block) {
			for _, sub := range subs {
				if sub != nil && !sub.Abandoned {
					switch sub.Sub.Filter.FeedName {
					case "blocks":
						sub.HandleOrAbandon(blockJson)
					case "transactions":
						if sub.Sub.Filter.AccountAddr != nil || len(sub.Sub.Filter.customFilters) > 0 {
							if trxJson := indexer.Filter(sub.Sub.Filter); len(trxJson) > 0 {
								sub.HandleOrAbandon(trxJson)
							}
						} else {
							sub.HandleOrAbandon(transactionsJson)
						}
					case "messages":
						if sub.Sub.Filter.AccountAddr != nil || len(sub.Sub.Filter.customFilters) > 0 {
							if msgJson := indexer.Filter(sub.Sub.Filter); len(msgJson) > 0 {
								sub.HandleOrAbandon(msgJson)
							}
						} else {
							sub.HandleOrAbandon(messagesJson)
						}
					}
				}
			}
		}
	}
	m.rw.RUnlock()

	return nil
}

func (m *SubManager) Add(handler *SubHandler) {
	m.rw.Lock()
	defer m.rw.Unlock()

	subs, ok := m.subs[handler.Sub.Filter]
	if !ok {
		subs = make([]*SubHandler, 0, 8)
	}

	subs = append(subs, handler)
	m.subs[handler.Sub.Filter] = subs
}

func (m *SubManager) Get(key Filter) (subs []*SubHandler, ok bool) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	subs, ok = m.subs[key]
	return
}

func (m *SubManager) GarbageCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			m.collectGarbage()
		case <-ctx.Done():
			return
		}
	}
}

func (m *SubManager) collectGarbage() {
	m.rw.Lock()
	defer m.rw.Unlock()

	for key, subs := range m.subs {
		newSubs := make([]*SubHandler, 0, 8)

		for _, v := range subs {
			if v != nil {
				if v.Abandoned {
					if err := v.Sub.Conn.Close(); err != nil {
						// TODO: handle error
					}
				} else {
					newSubs = append(newSubs, v)
				}
			}
		}

		if len(newSubs) == 0 {
			delete(m.subs, key)
		} else {
			m.subs[key] = newSubs
		}
	}
}

func blockToFeedBlock(block *ton.Block) *feed.BlockInFeed {
	return &feed.BlockInFeed{
		WorkchainId: block.Info.WorkchainId,
		Shard:       block.Info.Shard,
		SeqNo:       block.Info.SeqNo,
		Time:        uint64(time.Unix(int64(block.Info.GenUtime), 0).UTC().Unix()),
		StartLt:     block.Info.StartLt,
		RootHash:    block.Info.RootHash,
		FileHash:    block.Info.FileHash,

		TotalFeesNanograms: block.Info.ValueFlow.FeesCollected,
		TrxCount:           uint64(block.Info.BlockStats.TrxCount),
		ValueNanograms:     block.Info.BlockStats.SentNanograms,
	}
}

func trxToFeedTrx(trx *ton.Transaction) (*feed.TransactionInFeed, error) {
	var isTock uint8
	if trx.IsTock {
		isTock = 1
	}

	accountAddr := trx.AccountAddr
	var totalNanograms, totalFwdFeeNanograms, totalIhrFeeNanograms, totalImportFeeNanograms, msgInCreatedLt uint64
	var addrUf, msgInType, src, srcUf, dest, destUf string
	var srcWorkchainId, destWorkchainId int32
	var err error

	if trx.InMsg != nil {
		totalNanograms, totalFwdFeeNanograms, totalIhrFeeNanograms, totalImportFeeNanograms =
			trx.InMsg.ValueNanograms, trx.InMsg.FwdFeeNanograms, trx.InMsg.IhrFeeNanograms, trx.InMsg.ImportFeeNanograms

		src, dest = trx.InMsg.Src.Addr, trx.InMsg.Dest.Addr
		msgInCreatedLt = trx.InMsg.CreatedLt
		msgInType = trx.InMsg.Type

		if len(src) > 1 {
			if strings.HasPrefix(src, "x") {
				src = src[1:]
			}
		}
		if len(dest) > 1 {
			if strings.HasPrefix(src, "x") {
				dest = dest[1:]
			}
		}

		srcWorkchainId, destWorkchainId = trx.InMsg.Src.WorkchainId, trx.InMsg.Dest.WorkchainId

		if len(accountAddr) == 65 && strings.HasPrefix(accountAddr, "x") {
			accountAddr = accountAddr[1:]
		}

		if addrUf, err = utils.ComposeRawAndConvertToUserFriendly(trx.WorkchainId, accountAddr); err != nil {
			return nil, err
		}

		if len(src) == 64 {
			if srcUf, err = utils.ComposeRawAndConvertToUserFriendly(trx.WorkchainId, src); err != nil {
				return nil, err
			}
		}

		if len(dest) == 64 {
			if destUf, err = utils.ComposeRawAndConvertToUserFriendly(trx.WorkchainId, dest); err != nil {
				return nil, err
			}
		}
	}

	for _, msg := range trx.OutMsgs {
		totalNanograms += msg.ValueNanograms
		totalFwdFeeNanograms += msg.FwdFeeNanograms
		totalIhrFeeNanograms += msg.IhrFeeNanograms
		totalImportFeeNanograms += msg.ImportFeeNanograms
	}

	return &feed.TransactionInFeed{
		WorkchainId:   trx.WorkchainId,
		Shard:         trx.Shard,
		SeqNo:         trx.SeqNo,
		TimeUnix:      trx.Now,
		Lt:            trx.Lt,
		TrxHash:       trx.Hash,
		Type:          trx.Type,
		AccountAddr:   accountAddr,
		AccountAddrUF: addrUf,
		IsTock:        isTock,

		MsgInCreatedLt:  msgInCreatedLt,
		MsgInType:       msgInType,
		SrcWorkchainId:  srcWorkchainId,
		Src:             src,
		SrcUf:           srcUf,
		DestWorkchainId: destWorkchainId,
		Dest:            dest,
		DestUf:          destUf,

		TotalNanograms:          totalNanograms,
		TotalFeesNanograms:      trx.TotalFeesNanograms,
		TotalFwdFeeNanograms:    totalFwdFeeNanograms,
		TotalIhrFeeNanograms:    totalIhrFeeNanograms,
		TotalImportFeeNanograms: totalImportFeeNanograms,
	}, nil
}

func messageToFeedMessage(block *ton.Block, msg *ton.TransactionMessage, direction string, lt uint64) (*feed.MessageInFeed, error) {
	var src, dest = msg.Src.Addr, msg.Dest.Addr
	var srcUf, destUf, msgBody string
	var err error

	if len(src) > 1 && strings.HasPrefix(src, "x") {
		src = src[1:]
	}
	if len(dest) > 1 {
		dest = dest[1:]
	}
	if len(src) == 64 {
		if srcUf, err = utils.ComposeRawAndConvertToUserFriendly(msg.Src.WorkchainId, src); err != nil {
			return nil, err
		}
	}
	if len(dest) == 64 {
		if destUf, err = utils.ComposeRawAndConvertToUserFriendly(msg.Dest.WorkchainId, dest); err != nil {
			return nil, err
		}
	}

	// TODO: add support for x{00000001 format (encrypted) to all message body parsing
	if len(msg.BodyValue) >= 10 && msg.BodyValue[0:9] == "x{00000000" && msg.BodyValue != "x{00000000}" {
		replacer := strings.NewReplacer("x{", "", "}", "", "\t", "", "\n", "", " ", "")
		if msgBodyBytes, err := hex.DecodeString(replacer.Replace(msg.BodyValue)[8:]); err != nil {
			return nil, err
		} else {
			msgBody = string(msgBodyBytes)
		}
	}

	return &feed.MessageInFeed{
		WorkchainId: block.Info.WorkchainId,
		Shard:       block.Info.Shard,
		SeqNo:       block.Info.SeqNo,
		Lt:          lt,
		Time:        msg.CreatedAt,
		TrxHash:     msg.TrxHash,
		MessageLt:   msg.CreatedLt,
		Direction:   direction,

		SrcWorkchainId:  msg.Src.WorkchainId,
		Src:             src,
		SrcUf:           srcUf,
		DestWorkchainId: msg.Dest.WorkchainId,
		Dest:            dest,
		DestUf:          destUf,

		ValueNanogram:    msg.ValueNanograms,
		TotalFeeNanogram: msg.FwdFeeNanograms + msg.IhrFeeNanograms + msg.ImportFeeNanograms,
		Bounce:           msg.Bounce,
		Body:             msgBody,
	}, nil
}

func NewSubManager() *SubManager {
	return &SubManager{
		subs: make(map[Filter][]*SubHandler),
		rw:   sync.RWMutex{},
	}
}


package streaming

import (
	"encoding/hex"
	"strings"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

type FeedConverter interface {
	ConvertBlock(*ton.Block) (*feed.BlockInFeed, error)
	ConvertTransaction(*ton.Transaction) (*feed.TransactionInFeed, error)
	ConvertMessage(block *ton.Block, trx *ton.Transaction, msg *ton.TransactionMessage, direction string, lt uint64) (*feed.MessageInFeed, error)
}

type FeedConverterImpl struct{}

func (*FeedConverterImpl) ConvertBlock(block *ton.Block) (*feed.BlockInFeed, error) {
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
	}, nil
}

func (*FeedConverterImpl) ConvertTransaction(trx *ton.Transaction) (*feed.TransactionInFeed, error) {
	var isTock uint8
	if trx.IsTock {
		isTock = 1
	}

	accountAddr := trx.AccountAddr
	var totalNanograms, totalFwdFeeNanograms, totalIhrFeeNanograms, totalImportFeeNanograms, msgInCreatedLt uint64
	var addrUf, msgInType, src, srcUf, dest, destUf string
	var srcWorkchainId, destWorkchainId int32
	var err error

	if len(accountAddr) == 65 && strings.HasPrefix(accountAddr, "x") {
		accountAddr = accountAddr[1:]
	}

	if addrUf, err = utils.ComposeRawAndConvertToUserFriendly(trx.WorkchainId, accountAddr); err != nil {
		return nil, err
	}

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
			if strings.HasPrefix(dest, "x") {
				dest = dest[1:]
			}
		}

		srcWorkchainId, destWorkchainId = trx.InMsg.Src.WorkchainId, trx.InMsg.Dest.WorkchainId

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

func (*FeedConverterImpl) ConvertMessage(block *ton.Block, trx *ton.Transaction, msg *ton.TransactionMessage, direction string, lt uint64) (*feed.MessageInFeed, error) {
	var src, dest = msg.Src.Addr, msg.Dest.Addr
	var srcUf, destUf, msgBody string
	var err error

	if len(src) > 1 && strings.HasPrefix(src, "x") {
		src = src[1:]
	}
	if len(dest) > 1 && strings.HasPrefix(dest, "x") {
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

	var createdTime uint64
	if msg.CreatedAt != 0 {
		createdTime = msg.CreatedAt
	} else {
		createdTime = trx.Now
	}

	return &feed.MessageInFeed{
		WorkchainId: block.Info.WorkchainId,
		Shard:       block.Info.Shard,
		SeqNo:       block.Info.SeqNo,
		Lt:          lt,
		Time:        createdTime,
		TrxHash:     trx.Hash,
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

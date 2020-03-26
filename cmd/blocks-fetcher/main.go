package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gitlab.flora.loc/mills/tondb/internal/blocks_fetcher"
	"gitlab.flora.loc/mills/tondb/internal/tlb_pretty"
	"gitlab.flora.loc/mills/tondb/internal/ton"
)

func main() {
	blocksFetcher, err := blocks_fetcher.NewClient("127.0.0.1:13699")
	if err != nil {
		log.Fatal(err)
	}

	blockStr := "(-1,8000000000000000,958186)"
	blockId, err := ton.ParseBlockId(blockStr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", *blockId)

	tlbPrettyCustom, err := blocksFetcher.FetchBlockTlb(*blockId, blocks_fetcher.FormatPretty)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(tlbPrettyCustom))
	ioutil.WriteFile("/Users/user/go/src/github.com/yakud/tondb/bloc.tlb", tlbPrettyCustom, 0644)

	node := tlb_pretty.NewParser().Parse(tlbPrettyCustom)
	newNode, err := tlb_pretty.NewTreeSimplifier().Simplify(node)
	if err != nil {
		log.Fatal("Simplify err:", err)
	}

	valueFlow, err := extractValueFlow(newNode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", valueFlow)
}

func extractValueFlow(node *tlb_pretty.AstNode) (*ton.ValueFlow, error) {
	valueFlowRoot, err := node.GetNode("value_flow")
	if err != nil {
		return nil, err
	}
	valueFlow := &ton.ValueFlow{
		FromPrevBlk:  0,
		ToNextBlk:    0,
		Imported:     0,
		Exported:     0,
		FeesImported: 0,
		Recovered:    0,
		Created:      0,
		Minted:       0,
	}

	valueFlow.FeesCollected, _ = valueFlowRoot.GetUint64("fees_collected", "grams", "amount", "value")
	valueFlow.Exported, _ = valueFlowRoot.GetUint64("value_0", "exported", "grams", "amount", "value")
	valueFlow.FromPrevBlk, _ = valueFlowRoot.GetUint64("value_0", "from_prev_blk", "grams", "amount", "value")
	valueFlow.ToNextBlk, _ = valueFlowRoot.GetUint64("value_0", "to_next_blk", "grams", "amount", "value")
	valueFlow.Imported, _ = valueFlowRoot.GetUint64("value_0", "imported", "grams", "amount", "value")
	valueFlow.Created, _ = valueFlowRoot.GetUint64("value_1", "created", "grams", "amount", "value")
	valueFlow.FeesImported, _ = valueFlowRoot.GetUint64("value_1", "fees_imported", "grams", "amount", "value")
	valueFlow.Minted, _ = valueFlowRoot.GetUint64("value_1", "minted", "grams", "amount", "value")
	valueFlow.Recovered, _ = valueFlowRoot.GetUint64("value_1", "recovered", "grams", "amount", "value")

	return valueFlow, nil
}

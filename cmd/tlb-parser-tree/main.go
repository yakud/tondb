package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/tlb_pretty"
)

func PrintAst(node *tlb_pretty.AstNode, spacesCount int) {
	spaces := strings.Repeat(" ", spacesCount)
	for k, v := range node.Fields {
		if vv, ok := v.(string); ok {
			if len(vv) > 40 {
				fmt.Print(spaces, k, ": ", string(vv[:10]), "...\n")
			} else {
				fmt.Print(spaces, k, ": ", string(vv), "\n")
			}
		} else if vv, ok := v.(*tlb_pretty.AstNode); ok {
			fmt.Print(spaces, k, ": ", "........", "\n")
			if k == "extra" {
				PrintAst(vv, spacesCount+2)
			}
			if spacesCount > 0 {
				PrintAst(vv, spacesCount+2)
			}

			//PrintAst(vv, spacesCount+2)
		}
	}
}

func main() {
	// scp akisilev@46.4.4.150:/tmp/testlog.log* /tmp/
	input := flag.String("in", "/Users/user/go/src/github.com/yakud/ton-blocks-stream-receiver/sample.pretty", "input tlb pretty")
	out := flag.String("out", "/Users/user/go/src/github.com/yakud/ton-blocks-stream-receiver/sample_custom.json", "out json")
	//input := flag.String("in", "", "input tlb pretty")
	//out := flag.String("out", "", "out json")
	flag.Parse()

	data, err := ioutil.ReadFile(*input)
	if err != nil {
		log.Fatal(err)
	}

	// Parse
	since := time.Now()
	node := tlb_pretty.NewParser().Parse(data)
	//PrintAst(node, 0)
	fmt.Println("Parsed for:", time.Since(since))

	// Simplify
	since = time.Now()
	newNode, err := tlb_pretty.NewTreeSimplifier().Simplify(node)
	if err != nil {
		log.Fatal("Simplify err:", err)
	}
	fmt.Println("Simplified for:", time.Since(since))

	//since = time.Now()
	//block, err := tlb_pretty.NewAstTonConverter().ConvertToBlock(newNode)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	// since = time.Now()
	//state, err := tlb_pretty.NewAstTonConverter().ConvertToState(newNode)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Printf("%+v\n", state)
	//if err != nil {
	//	dd, err := newNode.ToJSON()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	if err := ioutil.WriteFile(*out, dd, 0600); err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	log.Fatal("converter err: ", errOrig)
	//}
	fmt.Println("Converted for:", time.Since(since))

	dd, _ := json.Marshal(newNode.PureFields())
	if err := ioutil.WriteFile(*out, dd, 0600); err != nil {
		log.Fatal(err)
	}
	//dd, _ := newNode.ToJSON()
	//if err := ioutil.WriteFile(*out, dd, 0600); err != nil {
	//	log.Fatal(err)
	//}
	//
	////////
	//dd, err = newNode.ToJSON()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//if err := ioutil.WriteFile(*out+".orig", dd, 0600); err != nil {
	//	log.Fatal(err)
	//}
}

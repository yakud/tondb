package streaming

import (
	"bufio"
	"encoding/json"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"log"
	"os"
	"strconv"
	"testing"
)

var index, _, transactions, messages = loadIndex()

func TestFetchTransactionsByAccountAddr(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(trx *feed.TransactionInFeed) bool {
		return trx.AccountAddr == addr
	}

	filter := Filter{
		FeedName: "transactions",
		AccountAddr: &fullAddr,
	}

	FetchTransactionsTest(t, cond, filter)
}

func TestFetchTransactionsByAccAddrAndTotalNanogramLt(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(trx *feed.TransactionInFeed) bool {
		return trx.AccountAddr == addr && trx.TotalNanograms < 2700000000
	}

	cf := CustomFilter{
		Field: FieldTrxTotalNanogam,
		Operation: OpLt,
		ValueString: "2700000000",
	}
	filter := Filter{
		FeedName: "transactions",
		AccountAddr: &fullAddr,
		CustomFilters: CustomFilters{cf},
	}

	FetchTransactionsTest(t, cond, filter)
}

func TestFetchTransactionsByAccAddrAndTotalNanogramGt(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(trx *feed.TransactionInFeed) bool {
		return trx.AccountAddr == addr && trx.TotalNanograms > 2700000000
	}

	cf := CustomFilter{
		Field: FieldTrxTotalNanogam,
		Operation: OpGt,
		ValueString: "2700000000",
	}
	filter := Filter{
		FeedName: "transactions",
		AccountAddr: &fullAddr,
		CustomFilters: CustomFilters{cf},
	}

	FetchTransactionsTest(t, cond, filter)
}

func TestFetchTransactionsByAccAddrAndTotalNanogramEq(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(trx *feed.TransactionInFeed) bool {
		return trx.AccountAddr == addr && trx.TotalNanograms == 2700000000
	}

	cf := CustomFilter{
		Field: FieldTrxTotalNanogam,
		Operation: OpEq,
		ValueString: "2700000000",
	}
	filter := Filter{
		FeedName: "transactions",
		AccountAddr: &fullAddr,
		CustomFilters: CustomFilters{cf},
	}

	FetchTransactionsTest(t, cond, filter)
}

func TestFetchTransactionsByAccAddrAndTotalNanogramRange(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(trx *feed.TransactionInFeed) bool {
		return trx.AccountAddr == addr && trx.TotalNanograms >= 2700000000 && trx.TotalNanograms <= 3200000000
	}

	cf := CustomFilter{
		Field: FieldTrxTotalNanogam,
		Operation: OpRange,
		ValueString: "[2700000000,3200000000]",
	}
	filter := Filter{
		FeedName: "transactions",
		AccountAddr: &fullAddr,
		CustomFilters: CustomFilters{cf},
	}

	FetchTransactionsTest(t, cond, filter)
}

func TestFetchTransactionsByTotalNanogramRange(t *testing.T) {
	cond := func(trx *feed.TransactionInFeed) bool {
		return trx.TotalNanograms >= 2700000000 && trx.TotalNanograms <= 3200000000
	}

	cf := CustomFilter{
		Field: FieldTrxTotalNanogam,
		Operation: OpRange,
		ValueString: "[2700000000,3200000000]",
	}
	filter := Filter{
		FeedName: "transactions",
		CustomFilters: CustomFilters{cf},
	}

	FetchTransactionsTest(t, cond, filter)
}

func TestFetchMessagesByAccountAddr(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(msg *feed.MessageInFeed) bool {
		return msg.Src == addr || msg.Dest == addr
	}

	filter := Filter{
		FeedName: "messages",
		AccountAddr: &fullAddr,
	}

	FetchMessagesTest(t, cond, filter)
}

func TestFetchMessagesByAccountAddrAndDirection(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(msg *feed.MessageInFeed) bool {
		return msg.Direction == "out" && (msg.Src == addr || msg.Dest == addr)
	}

	dir := MessageDirectionOut
	filter := Filter{
		FeedName: "messages",
		AccountAddr: &fullAddr,
		MessageDirection: &dir,
	}

	FetchMessagesTest(t, cond, filter)
}

func TestFetchMessagesByDirection(t *testing.T) {
	cond := func(msg *feed.MessageInFeed) bool {
		return msg.Direction == "out"
	}

	dir := MessageDirectionOut
	filter := Filter{
		FeedName: "messages",
		MessageDirection: &dir,
	}

	FetchMessagesTest(t, cond, filter)
}

func TestFetchMessagesByAccAddrAndValueNanogramsLt(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(msg *feed.MessageInFeed) bool {
		return (msg.Src == addr || msg.Dest == addr) && msg.ValueNanogram < 2700000000
	}

	cf := CustomFilter{
		Field: FieldMsgValueNanogam,
		Operation: OpLt,
		ValueString: "2700000000",
	}
	filter := Filter{
		FeedName: "messages",
		AccountAddr: &fullAddr,
		CustomFilters: CustomFilters{cf},
	}

	FetchMessagesTest(t, cond, filter)
}

func TestFetchMessagesByAccAddrAndTotalNanogramGt(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(msg *feed.MessageInFeed) bool {
		return (msg.Src == addr || msg.Dest == addr) && msg.ValueNanogram > 2700000000
	}

	cf := CustomFilter{
		Field: FieldMsgValueNanogam,
		Operation: OpGt,
		ValueString: "2700000000",
	}
	filter := Filter{
		FeedName: "messages",
		AccountAddr: &fullAddr,
		CustomFilters: CustomFilters{cf},
	}

	FetchMessagesTest(t, cond, filter)
}

func TestFetchMessagesByAccAddrAndTotalNanogramEq(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(msg *feed.MessageInFeed) bool {
		return (msg.Src == addr || msg.Dest == addr) && msg.ValueNanogram == 2700000000
	}

	cf := CustomFilter{
		Field: FieldMsgValueNanogam,
		Operation: OpEq,
		ValueString: "2700000000",
	}
	filter := Filter{
		FeedName: "messages",
		AccountAddr: &fullAddr,
		CustomFilters: CustomFilters{cf},
	}

	FetchMessagesTest(t, cond, filter)
}

func TestFetchMessagesByAccAddrAndTotalNanogramRange(t *testing.T) {
	addr := "3333333333333333333333333333333333333333333333333333333333333333"
	fullAddr := "-1:"+addr
	cond := func(msg *feed.MessageInFeed) bool {
		return (msg.Src == addr || msg.Dest == addr) && msg.ValueNanogram >= 2700000000 && msg.ValueNanogram <= 3205505017
	}

	cf := CustomFilter{
		Field: FieldMsgValueNanogam,
		Operation: OpRange,
		ValueString: "[2700000000, 3205505017]",
	}
	filter := Filter{
		FeedName: "messages",
		AccountAddr: &fullAddr,
		CustomFilters: CustomFilters{cf},
	}

	FetchMessagesTest(t, cond, filter)
}

func TestFetchMessagesByTotalNanogramRange(t *testing.T) {
	cond := func(msg *feed.MessageInFeed) bool {
		return msg.ValueNanogram >= 2700000000 && msg.ValueNanogram <= 3205505017
	}

	cf := CustomFilter{
		Field: FieldMsgValueNanogam,
		Operation: OpRange,
		ValueString: "[2700000000, 3205505017]",
	}
	filter := Filter{
		FeedName: "messages",
		CustomFilters: CustomFilters{cf},
	}

	FetchMessagesTest(t, cond, filter)
}

func TestFetchMessagesByDirectionAndTotalNanogramRange(t *testing.T) {
	dir := MessageDirectionOut
	cond := func(msg *feed.MessageInFeed) bool {
		return msg.Direction == string(dir) && msg.ValueNanogram >= 2700000000 && msg.ValueNanogram <= 3205505017
	}

	cf := CustomFilter{
		Field: FieldMsgValueNanogam,
		Operation: OpRange,
		ValueString: "[2700000000, 3205505017]",
	}
	filter := Filter{
		FeedName: "messages",
		CustomFilters: CustomFilters{cf},
		MessageDirection: &dir,
	}

	FetchMessagesTest(t, cond, filter)
}

func FetchTransactionsTest(t *testing.T, filterFunc func(*feed.TransactionInFeed) bool, filter Filter) {
	validTransactions := make(map[string]*feed.TransactionInFeed, 200)
	for _, trx := range transactions {
		if filterFunc(trx) {
			validTransactions[trx.TrxHash] = trx
		}
	}

	fetchedTransactions, err := index.FetchTransactions(filter)
	if err != nil {
		t.Error(err)
	}

	if len(validTransactions) != len(fetchedTransactions) {
		t.Errorf("len(validTransactions): %d, len(fetchedTransactions): %d", len(validTransactions), len(fetchedTransactions))
	}

	for _, fetched := range fetchedTransactions {
		if _, ok := validTransactions[fetched.TrxHash]; !ok {
			t.Error("no such trx in valid")
		}
	}
}

func FetchMessagesTest(t *testing.T, filterFunc func(*feed.MessageInFeed) bool, filter Filter) {
	validMessages := make(map[string]*feed.MessageInFeed, 200)
	for _, msg := range messages {
		if filterFunc(msg) {
			validMessages[msg.TrxHash+":"+strconv.FormatUint(msg.MessageLt,10)] = msg
		}
	}

	fetchedMessages, err := index.FetchMessage(filter)
	if err != nil {
		t.Error(err)
	}

	if len(validMessages) != len(fetchedMessages) {
		t.Errorf("len(validMessages): %d, len(fetchedMessages): %d", len(validMessages), len(fetchedMessages))
	}

	for _, fetched := range fetchedMessages {
		if _, ok := validMessages[fetched.TrxHash+":"+strconv.FormatUint(fetched.MessageLt,10)]; !ok {
			t.Error("no such msg in valid")
		}
	}
}

func loadIndex() (*Index, []*feed.BlockInFeed, []*feed.TransactionInFeed, []*feed.MessageInFeed) {
	index := NewIndex()
	transactions := make([]*feed.TransactionInFeed, 0, 200)
	messages := make([]*feed.MessageInFeed, 0, 200)
	blocks := make([]*feed.BlockInFeed, 0, 200)

	file, err := os.Open("test_files/blocks")

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		block := &feed.BlockInFeed{}
		if err = json.Unmarshal(scanner.Bytes(), block); err != nil {
			log.Fatal("couldn't unmarshal json")
		}
		if err = index.IndexBlock(block); err != nil {
			log.Fatal("couldn't index block")
		}

		blocks = append(blocks, block)

		// one block is enough so break
		break
	}
	file.Close()

	file, err = os.Open("test_files/transactions")

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	scanner = bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		transaction := &feed.TransactionInFeed{}
		if err = json.Unmarshal(scanner.Bytes(), transaction); err != nil {
			log.Fatal("couldn't unmarshal json")
		}
		if err = index.IndexTransaction(transaction); err != nil {
			log.Fatal("couldn't index transaction")
		}

		transactions = append(transactions, transaction)
	}
	file.Close()

	file, err = os.Open("test_files/messages")

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	scanner = bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		message := &feed.MessageInFeed{}
		if err = json.Unmarshal(scanner.Bytes(), message); err != nil {
			log.Fatal("couldn't unmarshal json")
		}
		if err = index.IndexMessage(message); err != nil {
			log.Fatal("couldn't index message")
		}

		messages = append(messages, message)
	}
	file.Close()

	return index, blocks, transactions, messages
}
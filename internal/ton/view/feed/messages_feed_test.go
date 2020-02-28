package feed

import (
	//"database/sql"
	"fmt"
	"testing"

	sql "github.com/jmoiron/sqlx"

	_ "github.com/mailru/go-clickhouse"
	"github.com/stretchr/testify/assert"
)

// Represent of original messages feed table
const testCreateMessagesTable = `CREATE TABLE IF NOT EXISTS transactions_test (
		WorkchainId           Int32,
		Shard                 UInt64,
		SeqNo                 UInt64,
		Time                  UInt64,
		Lt                    UInt64,
		MessageLt			  UInt64
	) ENGINE MergeTree
	PARTITION BY tuple()
	ORDER BY (Time, Lt, MessageLt, WorkchainId);
`

const testDropMessagesTable = `drop table transactions_test;`
const testInsertMessagesTable = `INSERT INTO transactions_test (WorkchainId,Shard,SeqNo,Time,Lt,MessageLt) VALUES (:WorkchainId,:Shard,:SeqNo,:Time,:Lt,:MessageLt)`

const testSelectAllMessages = `SELECT 
	WorkchainId, 
	Shard, 
	SeqNo, 
	Time, 
	Lt, 
	MessageLt 
FROM transactions_test
ORDER BY Time DESC, Lt DESC, MessageLt DESC, WorkchainId DESC`

const testSelectPartMessages = `
WITH (
	SELECT (min(Time), max(Time), max(Lt), max(MessageLt))
	FROM (
		SELECT 
		   Time,
		   Lt,
		   MessageLt
		FROM "transactions_test"
		PREWHERE
		     if(:time == 0, 1,
		        (Time = :time AND Lt <= :lt AND MessageLt < :message_lt) OR
		     	(Time < :time)
			 )
		ORDER BY Time DESC, Lt DESC, MessageLt DESC, WorkchainId DESC
		LIMIT :limit
	)
) as TimeRange
SELECT 
	WorkchainId, 
	Shard, 
	SeqNo, 
	Time, 
	Lt, 
	MessageLt 
FROM transactions_test
PREWHERE 
	 (Time >= TimeRange.1 AND Time <= TimeRange.2) AND
	 (Lt <= TimeRange.3 AND MessageLt <= TimeRange.4)
ORDER BY Time DESC, Lt DESC, MessageLt DESC, WorkchainId DESC
LIMIT :limit`

type testMessage struct {
	WorkchainId int32  `db:"WorkchainId"`
	Shard       uint64 `db:"Shard"`
	SeqNo       uint64 `db:"SeqNo"`
	Time        uint64 `db:"Time"`
	Lt          uint64 `db:"Lt"`
	MessageLt   uint64 `db:"MessageLt"`
}

type selectFilter struct {
	Time      uint64 `db:"time"`
	Lt        uint64 `db:"lt"`
	MessageLt uint64 `db:"message_lt"`
	Limit     uint64 `db:"limit"`
}

func TestMessagesFeed_SelectLatestMessages(t *testing.T) {
	conn, err := chConnection()
	assert.NoError(t, err)
	assert.NoError(t, testInitMessagesFeed(conn))

	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 5, Lt: 3475264000003, MessageLt: 3475264000002}))
	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 5, Lt: 3475264000001, MessageLt: 3475264000002}))
	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 5, Lt: 3475264000001, MessageLt: 3475259000002}))
	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 10, Lt: 1, MessageLt: 1}))
	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 20, Lt: 2, MessageLt: 2}))
	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 20, Lt: 2, MessageLt: 3}))
	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 20, Lt: 2, MessageLt: 4}))
	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 30, Lt: 3, MessageLt: 3}))
	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 30, Lt: 4, MessageLt: 4}))
	assert.NoError(t, testInsertMessageFeed(conn, &testMessage{Time: 30, Lt: 5, MessageLt: 5}))

	fmt.Println("ALL:")
	allMsgs, err := testSelectMessageFeed(conn, testSelectAllMessages, map[string]interface{}{})
	assert.NoError(t, err)
	for _, msg := range allMsgs {
		fmt.Printf("%+v\n", msg)
	}

	fmt.Println("First 0,1:")
	msgs, err := testSelectMessageFeed(conn, testSelectPartMessages, &selectFilter{
		Limit: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, allMsgs[0:2], msgs)
	for _, msg := range msgs {
		fmt.Printf("%+v\n", msg)
	}

	last := msgs[len(msgs)-1]

	fmt.Println("Next 2,3:")
	msgs, err = testSelectMessageFeed(conn, testSelectPartMessages, &selectFilter{
		Time:      last.Time,
		Lt:        last.Lt,
		MessageLt: last.MessageLt,
		Limit:     2,
	})
	assert.NoError(t, err)
	assert.Equal(t, allMsgs[2:4], msgs)
	for _, msg := range msgs {
		fmt.Printf("%+v\n", msg)
	}

	last = msgs[len(msgs)-1]
	fmt.Println("Next 4,5:")
	msgs, err = testSelectMessageFeed(conn, testSelectPartMessages, &selectFilter{
		Time:      last.Time,
		Lt:        last.Lt,
		MessageLt: last.MessageLt,
		Limit:     2,
	})
	assert.NoError(t, err)
	assert.Equal(t, allMsgs[4:6], msgs)
	for _, msg := range msgs {
		fmt.Printf("%+v\n", msg)
	}

	last = msgs[len(msgs)-1]
	fmt.Println("Next 6,7, 8:")
	msgs, err = testSelectMessageFeed(conn, testSelectPartMessages, &selectFilter{
		Time:      last.Time,
		Lt:        last.Lt,
		MessageLt: last.MessageLt,
		Limit:     3,
	})
	assert.NoError(t, err)
	assert.Equal(t, allMsgs[6:9], msgs)
	for _, msg := range msgs {
		fmt.Printf("%+v\n", msg)
	}

	last = msgs[len(msgs)-1]
	fmt.Println("Next 9,10,11:")
	msgs, err = testSelectMessageFeed(conn, testSelectPartMessages, &selectFilter{
		Time:      last.Time,
		Lt:        last.Lt,
		MessageLt: last.MessageLt,
		Limit:     3,
	})
	assert.NoError(t, err)
	assert.Equal(t, allMsgs[9:10], msgs)
	for _, msg := range msgs {
		fmt.Printf("%+v\n", msg)
	}
}

func testInsertMessageFeed(conn *sql.DB, messages ...*testMessage) error {
	for _, msg := range messages {
		_, err := conn.NamedExec(testInsertMessagesTable, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func testSelectMessageFeed(conn *sql.DB, query string, arg interface{}) ([]testMessage, error) {
	fmt.Printf("%+v\n", arg)
	rows, err := conn.NamedQuery(query, arg)
	if err != nil {
		return nil, err
	}
	var messages []testMessage
	for rows.Next() {
		msg := testMessage{}
		if err := rows.StructScan(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func testInitMessagesFeed(conn *sql.DB) error {
	_, _ = conn.Exec(testDropMessagesTable)
	_, err := conn.Exec(testCreateMessagesTable)
	return err
}

func chConnection() (*sql.DB, error) {
	connect, err := sql.Open("clickhouse", "http://127.0.0.1:8123/default?debug=false")
	if err != nil {
		return nil, fmt.Errorf("CH connect error: %s", err)
	}

	if err := connect.Ping(); err != nil {
		return nil, fmt.Errorf("CH ping error: %s", err)
	}

	return connect, nil
}

package index

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (
	createIndexHash = `
	CREATE TABLE IF NOT EXISTS _index_Hash (
	    Hash FixedString(64),
	    Type LowCardinality(String),
	    Data String
	) ENGINE MergeTree
	PARTITION BY tuple()
	ORDER BY (Hash)
	SETTINGS index_granularity = 64;
`
	// TODO: recreate with fill
	createBlocksHashIndex = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_index_HashBlock TO _index_Hash
	AS
	SELECT 
	   toFixedString(H.1, 64) as Hash,
	   toLowCardinality(H.2) as Type,
	   H.3 as Data
    FROM(
		SELECT 
		arrayJoin(
			[(
			   RootHash,
			   'block',
			   concat('(', toString(WorkchainId),',',hex(Shard),',',toString(SeqNo),')')
			),(
			   FileHash,
			   'block',
			   concat('(', toString(WorkchainId),',',hex(Shard),',',toString(SeqNo),')')
			)]
		) H
		FROM blocks 
	)
`
	// TODO: recreate with fill
	transactionHashIndex = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_index_HashTransaction TO _index_Hash
	AS
	SELECT 
	   Hash,
	   'transaction' as Type,
	   concat('(', toString(WorkchainId),',',hex(Shard),',',toString(SeqNo),')') as Data
    FROM transactions
`

	selectSomethingByHash = `
	SELECT 
		Hash,
		Type,
		Data
	FROM "_index_Hash"
	PREWHERE Hash = ?
	LIMIT 100
`

	dropIndexHash            = `DROP TABLE _index_Hash`
	dropBlocksHashIndex      = `DROP TABLE _view_index_HashBlock`
	dropTransactionHashIndex = `DROP TABLE _view_index_HashTransaction`
)

type SomethingType string

const (
	TypeBlock       SomethingType = "block"
	TypeTransaction SomethingType = "transaction"
	// todo add message
)

type Something struct {
	Hash string
	Type SomethingType
	Data string
}

type IndexHash struct {
	view.View
	conn *sql.DB
}

func (t *IndexHash) CreateTable() error {
	if _, err := t.conn.Exec(createIndexHash); err != nil {
		return err
	}
	if _, err := t.conn.Exec(createBlocksHashIndex); err != nil {
		return err
	}
	if _, err := t.conn.Exec(transactionHashIndex); err != nil {
		return err
	}

	return nil
}

func (t *IndexHash) DropTable() error {
	if _, err := t.conn.Exec(dropBlocksHashIndex); err != nil {
		return err
	}
	if _, err := t.conn.Exec(dropTransactionHashIndex); err != nil {
		return err
	}
	if _, err := t.conn.Exec(dropIndexHash); err != nil {
		return err
	}

	return nil
}

func (t *IndexHash) SelectSomethingByHash(hash string) ([]Something, error) {
	rows, err := t.conn.Query(selectSomethingByHash, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]Something, 0)
	for rows.Next() {
		something := Something{}
		err := rows.Scan(
			&something.Hash,
			&something.Type,
			&something.Data,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, something)
	}

	return res, nil
}

func NewIndexHash(conn *sql.DB) *IndexHash {
	return &IndexHash{
		conn: conn,
	}
}

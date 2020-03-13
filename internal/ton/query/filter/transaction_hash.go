package filter

type TransactionHashByIndex struct {
	hash string
}

func (f *TransactionHashByIndex) Build() (string, []interface{}, error) {
	filter := `((WorkchainId, Shard, SeqNo) IN (
		SELECT 
			WorkchainId, Shard, SeqNo 
		FROM ".inner._view_index_TransactionBlock"
		WHERE cityHash64(?) = Hash
	)) AND Hash = ?`
	args := []interface{}{f.hash, f.hash}

	return filter, args, nil
}

func NewTransactionHashByIndex(hash string) (*TransactionHashByIndex, error) {
	f := &TransactionHashByIndex{
		hash: hash,
	}

	return f, nil
}

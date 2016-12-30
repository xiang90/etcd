package backend

import (
	"bytes"

	"github.com/boltdb/bolt"
)

type ReadTx interface {
	Range(bucketName []byte, key, endKey []byte, limit int64) (keys [][]byte, vals [][]byte)
	Rollback() error
}

type readTx struct {
	tx *bolt.Tx
}

func (t *readTx) Range(bucketName []byte, key, endKey []byte, limit int64) (keys [][]byte, vs [][]byte) {
	bucket := t.tx.Bucket(bucketName)
	if bucket == nil {
		plog.Fatalf("bucket %s does not exist", bucketName)
	}

	if len(endKey) == 0 {
		if v := bucket.Get(key); v == nil {
			return keys, vs
		} else {
			return append(keys, key), append(vs, v)
		}
	}

	c := bucket.Cursor()
	for ck, cv := c.Seek(key); ck != nil && bytes.Compare(ck, endKey) < 0; ck, cv = c.Next() {
		vs = append(vs, cv)
		keys = append(keys, ck)
		if limit > 0 && limit == int64(len(keys)) {
			break
		}
	}

	return keys, vs
}

func (t *readTx) Rollback() error {
	return t.tx.Rollback()
}

package db

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"nicograshoff.de/x/optask/internal/model"
	"nicograshoff.de/x/optask/internal/stdstreams"
)

const OpenTimeout = 1 * time.Second

type Adapter struct {
	db *bolt.DB
	p  *model.Project
}

func NewAdapter(file string, p *model.Project) *Adapter {
	db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: OpenTimeout})
	if err != nil {
		log.Fatal(err)
	}

	err = updateSchema(db, p)
	if err != nil {
		log.Fatal(err)
	}

	return &Adapter{db, p}
}

func (a *Adapter) Close() {
	a.db.Close()
}

func (a *Adapter) CreateRun(tID model.TaskID, r *model.Run) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		bkt := taskRunBucket(tx, tID)

		s, _ := bkt.NextSequence()
		r.ID = model.RunID(itos(s))

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(r); err != nil {
			return err
		}
		b := buf.Bytes()

		if err := bkt.Put(itob(s), b); err != nil {
			return err
		}

		return nil
	})
}

func (a *Adapter) SaveRun(tID model.TaskID, r *model.Run) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return err
	}

	key, err := stob(string(r.ID))
	if err != nil {
		return err
	}

	return a.db.Update(func(tx *bolt.Tx) error {
		bkt := taskRunBucket(tx, tID)
		bkt.Put(key, buf.Bytes())
		return nil
	})
}

func (a *Adapter) LatestRuns() (map[model.TaskID]*model.Run, error) {
	ret := make(map[model.TaskID]*model.Run)
	err := a.db.View(func(tx *bolt.Tx) error {
		runsBkt := tx.Bucket([]byte("Runs"))
		tCur := runsBkt.Cursor()

		for tKey, _ := tCur.First(); tKey != nil; tKey, _ = tCur.Next() {
			tBkt := runsBkt.Bucket(tKey)
			rCur := tBkt.Cursor()
			rKey, rBin := rCur.Last()
			if rKey == nil {
				continue
			}

			r := model.Run{}
			buf := bytes.NewBuffer(rBin)
			dec := gob.NewDecoder(buf)
			if err := dec.Decode(&r); err != nil {
				return err
			}

			ret[model.TaskID(tKey)] = &r
		}

		return nil
	})
	return ret, err
}

func (a *Adapter) Runs(tID model.TaskID, before model.RunID, count int) ([]*model.Run, error) {
	ret := make([]*model.Run, count)
	err := a.db.View(func(tx *bolt.Tx) error {
		bkt := taskRunBucket(tx, tID)
		c := bkt.Cursor()

		var initCursor func() ([]byte, []byte)
		if before == "" {
			initCursor = func() ([]byte, []byte) { return c.Last() }
		} else {
			kb, err := stob(string(before))
			if err != nil {
				return err
			}

			initCursor = func() ([]byte, []byte) {
				c.Seek(kb)
				return c.Prev()
			}
		}

		n := 0
		for k, rBin := initCursor(); n < count; k, rBin = c.Prev() {
			if k == nil {
				small := make([]*model.Run, n)
				copy(small, ret)
				ret = small
				break
			}

			r := model.Run{}
			buf := bytes.NewBuffer(rBin)
			dec := gob.NewDecoder(buf)
			if err := dec.Decode(&r); err != nil {
				return err
			}

			ret[n] = &r
			n++
		}

		return nil
	})

	return ret, err
}

func (a *Adapter) Run(tID model.TaskID, rID model.RunID) (*model.Run, error) {
	k, err := stob(string(rID))
	if err != nil {
		return nil, err
	}

	var ret model.Run
	err = a.db.View(func(tx *bolt.Tx) error {
		bkt := taskRunBucket(tx, tID)
		data := bkt.Get(k)
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		return dec.Decode(&ret)
	})
	return &ret, err
}

func (a *Adapter) SaveLog(tID model.TaskID, rID model.RunID, l *stdstreams.Log) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(l); err != nil {
		return err
	}

	key, err := stob(string(rID))
	if err != nil {
		return err
	}

	return a.db.Update(func(tx *bolt.Tx) error {
		bkt := taskLogBucket(tx, tID)
		bkt.Put(key, buf.Bytes())
		return nil
	})
}

func (a *Adapter) Log(tID model.TaskID, rID model.RunID) (*stdstreams.Log, error) {
	key, err := stob(string(rID))
	if err != nil {
		return nil, err
	}

	var ret stdstreams.Log
	err = a.db.View(func(tx *bolt.Tx) error {
		bkt := taskLogBucket(tx, tID)
		data := bkt.Get(key)
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		return dec.Decode(&ret)
	})
	return &ret, err
}

func taskRunBucket(tx *bolt.Tx, tID model.TaskID) *bolt.Bucket {
	return tx.Bucket([]byte("Runs")).Bucket([]byte(tID))
}

func taskLogBucket(tx *bolt.Tx, tID model.TaskID) *bolt.Bucket {
	return tx.Bucket([]byte("Logs")).Bucket([]byte(tID))
}

func itos(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func itob(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func stob(s string) ([]byte, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return []byte{}, err
	}
	return itob(i), nil
}

// Package db provides types to access the persistence layer.
package db

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"nicograshoff.de/x/optask/internal/model"
	"nicograshoff.de/x/optask/internal/stdstreams"
)

const openTimeout = 1 * time.Second // timeout for bolt.Open
const dbPerm = 0600                 // file permissions for new databases

// Adapter represents a database adapter.
type Adapter struct {
	db *bolt.DB
	p  *model.Project
}

// NewAdapter creates an Adapter for the given database file. If the file does not exist, a
// dabase is created. Otherwise the database is opened and the schema is updated if necessary.
func NewAdapter(file string, p *model.Project) (*Adapter, error) {
	opts := bolt.Options{
		Timeout: openTimeout, // w/o Timeout bolt.Open blocks until file is unlocked
	}

	db, err := bolt.Open(file, dbPerm, &opts)
	if err != nil {
		return nil, err
	}

	err = updateSchema(db, p)
	if err != nil {
		return nil, err
	}

	return &Adapter{db, p}, nil
}

// Close closes the underlying database.
func (a *Adapter) Close() {
	a.db.Close()
}

// CreateRun saves the given run for the given task. Sets a task-unique run ID before persisting.
func (a *Adapter) CreateRun(tID model.TaskID, r *model.Run) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		bkt := taskRunBucket(tx, tID)

		s, _ := bkt.NextSequence()
		r.ID = model.RunID(itos(s))

		b, err := encRun(r)
		if err != nil {
			return err
		}

		if err := bkt.Put(itob(s), b); err != nil {
			return err
		}

		return nil
	})
}

// SaveRun saves the given run for the given task. Use CreateRun instead if the run does not have an ID yet.
func (a *Adapter) SaveRun(tID model.TaskID, r *model.Run) error {
	rBytes, err := encRun(r)
	if err != nil {
		return err
	}

	key, err := stob(string(r.ID))
	if err != nil {
		return err
	}

	return a.db.Update(func(tx *bolt.Tx) error {
		bkt := taskRunBucket(tx, tID)
		bkt.Put(key, rBytes)
		return nil
	})
}

// LatestRuns returns a map of model.TaskID mapped to the latest run each.
// If a task never ran, its ID will not be in the map.
func (a *Adapter) LatestRuns() (map[model.TaskID]*model.Run, error) {
	ret := make(map[model.TaskID]*model.Run)
	err := a.db.View(func(tx *bolt.Tx) error {
		runsBkt := tx.Bucket([]byte("Runs"))
		tCur := runsBkt.Cursor()

		for tKey, _ := tCur.First(); tKey != nil; tKey, _ = tCur.Next() {
			tBkt := runsBkt.Bucket(tKey)
			rCur := tBkt.Cursor()
			rKey, b := rCur.Last()
			if rKey == nil {
				continue // never ran
			}

			r, err := decRun(b)
			if err != nil {
				return err
			}

			ret[model.TaskID(tKey)] = r
		}
		return nil
	})
	return ret, err
}

// Runs returns count runs for a given task ordered by time of creation. If before is not empty,
// only runs that where created before that given run are returned.
func (a *Adapter) Runs(tID model.TaskID, before model.RunID, count int) ([]*model.Run, error) {
	ret := make([]*model.Run, count)
	err := a.db.View(func(tx *bolt.Tx) error {
		bkt := taskRunBucket(tx, tID)
		c := bkt.Cursor()

		var rKey, rBin []byte
		if before == "" {
			rKey, rBin = c.Last()
		} else {
			kb, err := stob(string(before))
			if err != nil {
				return err
			}

			c.Seek(kb)
			rKey, rBin = c.Prev()
		}

		n := 0
		for ; n < count; rKey, rBin = c.Prev() {
			if rKey == nil {
				small := make([]*model.Run, n)
				copy(small, ret)
				ret = small
				break
			}

			r, err := decRun(rBin)
			if err != nil {
				return err
			}

			ret[n] = r
			n++
		}

		return nil
	})

	return ret, err
}

// Run returns a pointer to a mode.Run for the task with the given id.
func (a *Adapter) Run(tID model.TaskID, rID model.RunID) (*model.Run, error) {
	k, err := stob(string(rID))
	if err != nil {
		return nil, err
	}

	var ret *model.Run
	err = a.db.View(func(tx *bolt.Tx) error {
		bkt := taskRunBucket(tx, tID)
		data := bkt.Get(k)
		ret, err = decRun(data)
		return err
	})
	return ret, err
}

// SaveLog persists a stdstreams.Log for a given task and run.
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

// Log returns a pointer to a stdstreams.Log associated with the given task and run.
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

func encRun(r *model.Run) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decRun(b []byte) (*model.Run, error) {
	r := model.Run{}
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
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

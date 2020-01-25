package db

import (
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"nicograshoff.de/x/optask/internal/model"
	"nicograshoff.de/x/optask/internal/stdstreams"
)

const TasksBucket = "Tasks"
const LogKey = "Log"

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

func (a *Adapter) CreateRun(tID model.TaskID) (model.RunID, error) {
	var rID model.RunID

	err := a.db.Update(func(tx *bolt.Tx) error {
		tasks := tx.Bucket([]byte(TasksBucket))
		task := tasks.Bucket([]byte(tID))
		id, _ := task.NextSequence()
		rID = model.RunID(itos(id))
		task.CreateBucket([]byte(rID))

		return nil
	})

	return rID, err
}

func (a *Adapter) LatestRun(tID model.TaskID) (model.RunID, error) {
	var rID model.RunID

	err := a.db.View(func(tx *bolt.Tx) error {
		task := taskBucket(tx, tID)
		c := task.Cursor()
		key, _ := c.Last()

		rID = model.RunID(key)
		return nil
	})

	return rID, err
}

func (a *Adapter) SaveRunLog(tID model.TaskID, rID model.RunID, log *stdstreams.Log) error {
	err := a.db.Update(func(tx *bolt.Tx) error {
		run := runBucket(tx, tID, rID)

		binLog, err := log.MarshalBinary()
		if err != nil {
			return err
		}

		run.Put([]byte(LogKey), binLog)
		return nil
	})

	return err
}

func (a *Adapter) RunLog(tID model.TaskID, rID model.RunID) (*stdstreams.Log, error) {
	var log *stdstreams.Log
	err := a.db.View(func(tx *bolt.Tx) error {
		run := runBucket(tx, tID, rID)
		binLog := run.Get([]byte(LogKey))

		log = stdstreams.NewLog()
		err := log.UnmarshalBinary(binLog)
		if err != nil {
			log = nil
		}
		return err
	})

	return log, err
}

func taskBucket(tx *bolt.Tx, tID model.TaskID) *bolt.Bucket {
	return tx.Bucket([]byte(TasksBucket)).Bucket([]byte(tID))
}

func runBucket(tx *bolt.Tx, tID model.TaskID, rID model.RunID) *bolt.Bucket {
	return taskBucket(tx, tID).Bucket([]byte(rID))
}

func itos(i uint64) string {
	return strconv.FormatUint(i, 10)
}

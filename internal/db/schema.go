package db

import (
	"github.com/boltdb/bolt"
	"nicograshoff.de/x/optask/internal/model"
)

func updateSchema(db *bolt.DB, p *model.Project) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Create Tasks bucket
		tasks, err := tx.CreateBucketIfNotExists([]byte(TasksBucket))
		if err != nil {
			return err
		}

		// Create bucket per task
		for _, task := range p.Tasks {
			_, err = tasks.CreateBucketIfNotExists([]byte(task.ID))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

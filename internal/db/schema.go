package db

import (
	"github.com/boltdb/bolt"
	"github.com/ngrash/optask/internal/model"
)

func updateSchema(db *bolt.DB, p *model.Project) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Create Runs bucket
		rBkt, err := tx.CreateBucketIfNotExists([]byte("Runs"))
		if err != nil {
			return err
		}

		// Create Logs bucket
		lBkt, err := tx.CreateBucketIfNotExists([]byte("Logs"))
		if err != nil {
			return nil
		}

		// Create bucket per task
		for _, t := range p.Tasks {
			_, err = rBkt.CreateBucketIfNotExists([]byte(t.ID))
			if err != nil {
				return err
			}

			_, err = lBkt.CreateBucketIfNotExists([]byte(t.ID))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

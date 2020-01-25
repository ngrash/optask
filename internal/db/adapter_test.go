package db

import (
	"io/ioutil"
	"os"
	"testing"

	"nicograshoff.de/x/optask/internal/model"
	"nicograshoff.de/x/optask/internal/stdstreams"
)

var project = &model.Project{
	"testing",
	"Testing",
	[]model.Task{
		model.Task{"t1", "Task 1", "true", []string{}},
		model.Task{"t2", "Task 2", "false", []string{}},
	},
}

func withTmpDB(t *testing.T, fn func(*Adapter)) {
	f, err := ioutil.TempFile("", "optask-testing.*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	defer os.Remove(f.Name())

	db := NewAdapter(f.Name(), project)
	defer db.Close()

	fn(db)
}

func TestCreateRun(t *testing.T) {
	withTmpDB(t, func(a *Adapter) {
		rID, err := a.CreateRun(project.Tasks[0].ID)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if rID != "1" {
			t.Errorf("Expected rID == 1, got: %v", rID)
		}
	})
}

func TestLatestRun(t *testing.T) {
	withTmpDB(t, func(a *Adapter) {
		tID := project.Tasks[0].ID
		a.CreateRun(tID)
		rID, _ := a.CreateRun(tID)

		if latest, _ := a.LatestRun(tID); latest != rID {
			t.Errorf("Expected latest == %v, got: %v", rID, latest)
		}
	})
}

func TestRunLog(t *testing.T) {
	withTmpDB(t, func(a *Adapter) {
		tID := project.Tasks[0].ID
		rID, _ := a.CreateRun(tID)
		log := stdstreams.NewLog()
		log.Stdout().Write([]byte("hello\n"))
		err := a.SaveRunLog(tID, rID, log)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		log2, err := a.RunLog(tID, rID)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		} else if log2 == nil {
			t.Fatal("err == nil but log == nil too")
		}

		if len(log2.Lines()) != 1 {
			t.Errorf("Expected 1 line, got: %v", len(log2.Lines()))
		}

		if log2.Lines()[0].Text != "hello" {
			t.Errorf("Expected Text == \"hello\", got: \"%v\"", log2.Lines()[0].Text)
		}
	})
}

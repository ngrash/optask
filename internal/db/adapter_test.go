package db

import (
	"io/ioutil"
	"os"
	"testing"

	"nicograshoff.de/x/optask/internal/model"
	"nicograshoff.de/x/optask/internal/stdstreams"
)

var project = &model.Project{
	ID:   "testing",
	Name: "Testing",
	Tasks: []model.Task{
		model.Task{ID: "t1", Name: "Task 1", Cmd: "true", Args: []string{}},
		model.Task{ID: "t2", Name: "Task 2", Cmd: "false", Args: []string{}},
	},
}

func TestCreateRun(t *testing.T) {
	withTmpDB(t, func(a *Adapter) {
		r := model.Run{}
		err := a.CreateRun(project.Tasks[0].ID, &r)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if r.ID != "1" {
			t.Errorf("Expected rID == 1, got: %v", r.ID)
		}
	})
}

func TestSaveRun(t *testing.T) {
	withTmpDB(t, func(a *Adapter) {
		r := model.Run{ID: "1"}
		err := a.SaveRun(project.Tasks[0].ID, &r)
		if err != nil {
			t.Fatalf("Unexpected error. %v", err)
		}
	})
}

func TestLatestRuns(t *testing.T) {
	withTmpDB(t, func(a *Adapter) {
		var r model.Run

		tID1 := project.Tasks[0].ID
		tID2 := project.Tasks[1].ID

		a.CreateRun(tID1, &r)
		a.CreateRun(tID1, &r)

		runs, err := a.LatestRuns()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if runs[tID1].ID != "2" {
			t.Errorf("Expected ID == \"2\", got: \"%v\"", runs[tID1].ID)
		}

		if runs[tID2] != nil {
			t.Errorf("Expected runs[tID2] == nil but wasn't")
		}
	})
}

func TestRun(t *testing.T) {
	withTmpDB(t, func(a *Adapter) {
		var r model.Run
		tID := project.Tasks[0].ID
		a.CreateRun(tID, &r)

		subject, err := a.Run(tID, r.ID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if subject.ID != r.ID {
			t.Errorf("Expected ID == \"%v\", got: \"%v\"", r.ID, subject.ID)
		}
	})
}

func TestRuns(t *testing.T) {
	t.Run("Order", func(t *testing.T) {
		withTmpDB(t, func(a *Adapter) {
			tID := project.Tasks[0].ID

			var r model.Run
			a.CreateRun(tID, &r)
			a.CreateRun(tID, &r)
			a.CreateRun(tID, &r)

			runs, err := a.Runs(tID, "", 3)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(runs) != 3 {
				t.Fatalf("Expected 3 runs, got: %v", len(runs))
			}

			if runs[0].ID != "3" {
				t.Errorf("Expected runs[0].ID == 3, got: %v", runs[0].ID)
			}

			if runs[1].ID != "2" {
				t.Errorf("Expected runs[1].ID == 3, got: %v", runs[1].ID)
			}

			if runs[2].ID != "1" {
				t.Errorf("Expected runs[2].ID == 3, got: %v", runs[2].ID)
			}
		})
	})

	t.Run("LessRunsThanCount", func(t *testing.T) {
		withTmpDB(t, func(a *Adapter) {
			tID := project.Tasks[0].ID

			var r model.Run
			a.CreateRun(tID, &r) // "1"
			a.CreateRun(tID, &r) // "2"
			a.CreateRun(tID, &r) // "3"

			runs, err := a.Runs(tID, "2", 10)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(runs) != 1 {
				t.Errorf("Expected 1 run, got: %v", len(runs))
			}
		})
	})
}

func TestSaveLog(t *testing.T) {
	withTmpDB(t, func(a *Adapter) {
		l := stdstreams.NewLog()
		l.Stdout().Write([]byte("hello\n"))

		err := a.SaveLog(project.Tasks[0].ID, "1", l)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

func TestLog(t *testing.T) {
	withTmpDB(t, func(a *Adapter) {
		tID := project.Tasks[0].ID

		l := stdstreams.NewLog()
		l.Stdout().Write([]byte("hello\n"))
		a.SaveLog(tID, "1", l)

		l, err := a.Log(tID, "1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if l.Lines()[0].Text != "hello" {
			t.Errorf("Expected Text == \"hello\", got: \"%v\"", l.Lines()[0].Text)
		}
	})
}

func withTmpDB(t *testing.T, fn func(*Adapter)) {
	f, err := ioutil.TempFile("", "optask-testing.*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	defer os.Remove(f.Name())

	db, err := NewAdapter(f.Name(), project)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer db.Close()

	fn(db)
}

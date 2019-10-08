package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FileSystem struct {
	root string
	mu   *sync.Mutex
}

type Node struct {
	path string
}

func NewFileSystem(root string) *FileSystem {
	err := os.MkdirAll(root, os.ModePerm|os.ModeDir)
	if err != nil {
		panic(err)
	}
	return &FileSystem{root, new(sync.Mutex)}
}

func (fs *FileSystem) NodeID(n *Node) string {
	return strings.ReplaceAll(n.path[len(fs.root):], "/", "")
}

func (fs *FileSystem) Node(id string) *Node {
	t, err := time.Parse("20060102", id[0:8])
	if err != nil {
		panic(err)
	}
	index, err := strconv.Atoi(id[8:])

	path := fs.nodePath(t, index)

	return &Node{path}
}

func (fs *FileSystem) Create(n *Node, name string) (*os.File, error) {
	return os.Create(path.Join(n.path, name))
}

func (fs *FileSystem) Open(n *Node, name string) (*os.File, error) {
	return os.Open(path.Join(n.path, name))
}

func (fs *FileSystem) nodesPath(t time.Time) string {
	p := fmt.Sprintf("%04d/%02d/%02d", t.Year(), t.Month(), t.Day())
	return path.Join(fs.root, p)
}

func (fs *FileSystem) nodePath(t time.Time, i int) string {
	nodesPath := fs.nodesPath(t)
	node := fmt.Sprintf("%d", i)
	return path.Join(nodesPath, node)
}

func (fs *FileSystem) CreateNodeNow() *Node {
	return fs.CreateNode(time.Now())
}

func (fs *FileSystem) CreateNode(t time.Time) *Node {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nodes := fs.ListNodes(t)
	index := len(nodes)
	path := fs.nodePath(t, index)
	os.MkdirAll(path, os.ModePerm|os.ModeDir)

	return &Node{path}
}

func (fs *FileSystem) LatestNode() *Node {
	// Relies on the sorted nature of ioutil.ReadDir for finding the
	// latest year, month, and day.
	years, err := ioutil.ReadDir(fs.root)
	if err != nil {
		panic(err)
	}

	// This is the only time we need to check if a directory is empty.
	// Directories are only created when nodes are created, so if there
	// is a year there is also a month, a day and an index.
	if len(years) <= 0 {
		return nil
	}

	// Relies on file system sorting. Breaks if used before
	// and after 9999 A.D.
	year := years[len(years)-1].Name()
	yearPath := path.Join(fs.root, year)

	// Relies on file system sorting. Not expected to break as months
	// are used with a leading zero (e.g. "02") and will not exceed 12.
	months, err := ioutil.ReadDir(yearPath)
	if err != nil {
		panic(err)
	}
	month := months[len(months)-1].Name()
	monthPath := path.Join(yearPath, month)

	// Relies on file system sorting. Again this is not a problem as days
	// are used with a leading zero (e.g. "08") and will not exceed 31.
	days, err := ioutil.ReadDir(monthPath)
	if err != nil {
		panic(err)
	}
	day := days[len(days)-1].Name()
	dayPath := path.Join(monthPath, day)

	// Indices are a special because they are not zero-extended. That is
	// because zero-extending them would not solve the problem but only
	// delay it. Maybe it is worth the effort one day. For now we can rely
	// on file system sorting for up to 10 indices ("0" to "9".) After that
	// we have to check their lenth and compare length (see code below.)
	indices, err := ioutil.ReadDir(dayPath)
	if err != nil {
		panic(err)
	}
	if len(indices) <= 10 {
		index := indices[len(indices)-1].Name()
		return &Node{path.Join(dayPath, index)}
	} else {
		maxIndex := ""
		for _, indexPath := range indices {
			index := indexPath.Name()
			// longer string -> bigger number
			if len(index) > len(maxIndex) {
				maxIndex = index
			} else {
				// same length -> compare
				if len(index) == len(maxIndex) && index > maxIndex {
					maxIndex = index
				}
			}
		}

		return &Node{path.Join(dayPath, maxIndex)}
	}
}

func (fs *FileSystem) ListNodes(time time.Time) []Node {
	nodesPath := fs.nodesPath(time)
	_, err := os.Stat(nodesPath)
	if err == nil {
		files, err := ioutil.ReadDir(nodesPath)
		if err != nil {
			panic(err)
		}
		nodes := make([]Node, len(files))
		for index, file := range files {
			nodes[index] = Node{path.Join(nodesPath, file.Name())}
		}
		return nodes
	} else if os.IsNotExist(err) {
		return make([]Node, 0)
	} else {
		panic(err)
	}
}

func (fs *FileSystem) ListNodesAfter(time time.Time) []Node {
	nodes := make([]Node, 0)
	err := filepath.Walk(fs.root, func(filePath string, info os.FileInfo, err error) error {
		parts := strings.Split(filePath, "/")

		// If the year is in the future we don't have to check for the months or days
		checkNext := true

		if len(parts) >= 2 {
			itemYear, err := strconv.Atoi(parts[1])
			if err != nil {
				panic(err)
			}

			yearLimit := time.Year()
			if itemYear < yearLimit {
				return filepath.SkipDir
			} else if itemYear > yearLimit {
				checkNext = false
			}
		}

		if checkNext && len(parts) >= 3 {
			itemMonth, err := strconv.Atoi(parts[2])
			if err != nil {
				panic(err)
			}

			monthLimit := int(time.Month())
			if itemMonth < monthLimit {
				return filepath.SkipDir
			} else if itemMonth > monthLimit {
				checkNext = false
			}
		}

		if checkNext && len(parts) >= 4 {
			day, err := strconv.Atoi(parts[3])
			if err != nil {
				panic(err)
			}
			if day < time.Day() {
				return filepath.SkipDir
			}
		}

		if len(parts) >= 5 {
			_, err := strconv.Atoi(parts[4])
			if err != nil {
				panic(err)
			}

			nodes = append(nodes, Node{filePath})

			// we are not interested in the node's content
			return filepath.SkipDir
		}

		// everything is fine, continue with next directory
		return nil
	})
	if err != nil {
		panic(err)
	}
	return nodes
}

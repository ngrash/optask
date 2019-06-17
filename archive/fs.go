package archive

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileSystem struct {
	root string
}

func NewFileSystem(root string) *FileSystem {
	err := os.MkdirAll(root, os.ModePerm|os.ModeDir)
	if err != nil {
		log.Panic(err)
	}
	return &FileSystem{root}
}

func (fs *FileSystem) Create(node *Node, name string) (*os.File, error) {
	return os.Create(fs.path(node, name))
}

func (fs *FileSystem) Open(node *Node, name string) (*os.File, error) {
	return os.Open(fs.path(node, name))
}

func (fs *FileSystem) CreateNode(id Identifier) *Node {
	idPath := path.Join(fs.root, id.path)
	os.MkdirAll(idPath, os.ModePerm|os.ModeDir)

	nodes := fs.ListNodes(id)
	nextIndex := len(nodes)
	newNode := &Node{id, nextIndex}
	newNodePath := path.Join(idPath, strconv.Itoa(newNode.index))
	os.MkdirAll(newNodePath, os.ModePerm|os.ModeDir)
	return newNode
}

func (fs *FileSystem) ListNodes(id Identifier) []Node {
	idPath := path.Join(fs.root, id.path)
	files, err := ioutil.ReadDir(idPath)
	if err != nil {
		log.Panic(err)
	}
	nodes := make([]Node, len(files))
	for index, _ := range files {
		nodes[index] = Node{id, index}
	}
	return nodes
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
				log.Panic(err)
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
				log.Panic(err)
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
				log.Panic(err)
			}
			if day < time.Day() {
				return filepath.SkipDir
			}
		}

		if len(parts) >= 5 {
			index, err := strconv.Atoi(parts[4])
			if err != nil {
				log.Panic(err)
			}

			id := Identifier{path.Join(parts[1:4]...)}
			nodes = append(nodes, Node{id, index})

			// we are not interested in the node's content
			return filepath.SkipDir
		}

		// everything is fine, continue with next directory
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return nodes
}

func (fs *FileSystem) path(node *Node, name string) string {
	return path.Join(fs.root, node.id.path, strconv.Itoa(node.index), name)
}

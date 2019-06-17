package archive

import (
	"log"
	"path"
	"strconv"
	"strings"
)

type Node struct {
	id    Identifier
	index int
}

func ParseNode(s string) Node {
	idPath := path.Join(s[0:4], s[4:6], s[6:8])
	id := Identifier{idPath}
	index, err := strconv.Atoi(s[8:])
	if err != nil {
		log.Panic(err)
	}

	return Node{id, index}
}

func (node Node) String() string {
	return strings.ReplaceAll(node.id.path, "/", "") + strconv.Itoa(node.index)
}

package archive

import (
	"fmt"
	"time"
)

type Identifier struct {
	path string
}

func NewIdentifierNow() Identifier {
	return NewIdentifier(time.Now())
}

func NewIdentifier(time time.Time) Identifier {
	path := fmt.Sprintf("%04d/%02d/%02d", time.Year(), time.Month(), time.Day())
	return Identifier{path}
}

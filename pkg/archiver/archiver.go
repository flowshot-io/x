package archiver

import (
	"github.com/mholt/archiver/v3"
)

type Archiver struct{}

func (a *Archiver) Archive(sources []string, destination string) error {
	return archiver.Archive(sources, destination)
}

func (a *Archiver) Unarchive(source string, destination string) error {
	return archiver.Unarchive(source, destination)
}

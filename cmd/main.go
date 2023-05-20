package main

import (
	"fmt"

	"github.com/flowshot-io/x/pkg/artifact"
)

func main() {
	err := artifactExamples()
	if err != nil {
		panic(err)
	}
}

func artifactExamples() error {
	err := directoryExample()
	if err != nil {
		return fmt.Errorf("error running directory example: %w", err)
	}

	err = fileExample()
	if err != nil {
		return fmt.Errorf("error running file example: %w", err)
	}

	return nil
}

func directoryExample() error {
	artifact, err := artifact.NewWithPaths("test", []string{"./pkg"})
	if err != nil {
		return err
	}

	fmt.Println("Artifact created:", artifact.GetName())

	files, err := artifact.ListFiles()
	if err != nil {
		return err
	}

	fmt.Println("Files:", files)

	return nil
}

func fileExample() error {
	artifact, err := artifact.NewWithPaths("test", []string{"README.md"})
	if err != nil {
		return err
	}

	fmt.Println("Artifact created:", artifact.GetName())

	files, err := artifact.ListFiles()
	if err != nil {
		return err
	}

	fmt.Println("Files:", files)

	return nil
}

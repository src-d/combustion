package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/src-d/combustion"
)

const AppName = "combustion"

func main() {
	parser := flags.NewNamedParser(AppName, flags.Default)
	parser.Command, _ = parser.AddCommand(AppName, "", "Options:", &Command{})

	_, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrCommandRequired {
			parser.WriteHelp(os.Stdout)
		}

		os.Exit(1)
	}
}

type Command struct {
	Output string `short:"o" long:"output" description:"output folder"`
	Input  struct {
		Folders []string `positional-arg-name:"input" description:"List of folder to process"`
	} `positional-args:"yes"`

	files []string
}

func (c *Command) Execute(args []string) error {
	if err := c.findAllFiles(); err != nil {
		return err
	}

	for _, file := range c.files {
		c.render(file)
	}

	fmt.Println(c.files)
	return nil
}

func (c *Command) findAllFiles() error {
	if err := c.findFiles("/*.yaml"); err != nil {
		return err
	}

	if err := c.findFiles("/**/*.yaml"); err != nil {
		return err
	}

	return nil
}

func (c *Command) findFiles(pattern string) error {
	for _, folder := range c.Input.Folders {
		results, err := filepath.Glob(folder + pattern)
		if err != nil {
			return err
		}

		c.files = append(c.files, results...)
	}

	return nil
}

func (c *Command) render(file string) error {
	cfg, err := combustion.NewConfigFromFile(file, nil)
	if err != nil {
		return err
	}

	r, err := cfg.SaveTo(c.Output)
	fmt.Println(r)
	if err != nil {
		return err
	}

	return err
}

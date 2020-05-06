package main

import (
	"flag"
	"github.com/DaniilDenisyuk/design-practice-2/build/gomodule"
	"github.com/DaniilDenisyuk/design-practice-2/build/zip_archive"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var (
	dryRun  = flag.Bool("dry-run", false, "Generate ninja build file but don't start the build")
	verbose = flag.Bool("v", false, "Display debugging logs")
)

func NewContext() *blueprint.Context {
	ctx := bood.PrepareContext()
	ctx.RegisterModuleType("zip_archive", zip_archive.SimpleArchiveFactory)
	ctx.RegisterModuleType("go_binary", gomodule.SimpleBinFactory)
	return ctx
}

func main() {
	flag.Parse()

	config := bood.NewConfig()
	if !*verbose {
		config.Debug = log.New(ioutil.Discard, "", 0)
	}
	ctx := NewContext()

	ninjaBuildPath := bood.GenerateBuildFile(config, ctx)

	if !*dryRun {
		config.Info.Println("Starting the build now")

		cmd := exec.Command("ninja", append([]string{"-f", ninjaBuildPath}, flag.Args()...)...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			config.Info.Fatal("Error invoking ninja build. See logs above.")
		}
	}
}

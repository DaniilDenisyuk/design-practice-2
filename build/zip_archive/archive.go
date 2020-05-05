package zip_archive

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
	"strings"
)

var (
	// Package context used to define Ninja build rules.
	pctx = blueprint.NewPackageContext("github.com/DaniilDenysiuk/design-practice-2/build/zip_archive")

	// Ninja rule to archive binary output file.
	makeArchive = pctx.StaticRule("makeArchive", blueprint.RuleParams{
		Command:     "cd $workDir && zip $outputPath -j $inputFiles",
		Description: "make archive from $inputFiles ",
	}, "workDir", "outputPath", "inputFiles")
)

type zipArchiveModule struct {
	blueprint.SimpleName

	properties struct {
		// Go binary filename to archive
		Srcs        []string
		SrcsExclude []string
	}
}

func (gb *zipArchiveModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	archiveName := name + ".zip"
	config := bood.ExtractConfig(ctx)
	outputPath := path.Join(config.BaseOutputDir, "archives", archiveName)
	var inputs []string
	inputErors := false
	for _, src := range gb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, gb.properties.SrcsExclude); err == nil {
			inputs = append(inputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErors = true
		}
	}
	if inputErors {
		return
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as zip archive", name),
		Rule:        makeArchive,
		Outputs:     []string{outputPath},
		Implicits:   nil,
		Args: map[string]string{
			"workDir":    ctx.ModuleDir(),
			"outputPath": outputPath,
			"inputFiles": strings.Join(inputs, ","),
		},
	})
}

// SimpleArchiveFactory is a factory for go binary module type which supports Go command packages with running tests.
func SimpleArchiveFactory() (blueprint.Module, []interface{}) {
	mType := &zipArchiveModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}

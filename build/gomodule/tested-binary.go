package gomodule

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
	"strings"
)

var (
	// Package context used to define Ninja build rules.
	pctx = blueprint.NewPackageContext("github.com/DaniilDenysiuk/design-practice-2/build/gomodule")

	// Ninja rule to execute go test.
	goTest = pctx.StaticRule("gotest", blueprint.RuleParams{
		Command:     "cd $workDir && go test -v $testPkg > $outReportPath",
		Description: "test $testPkg",
	}, "workDir", "testPkg", "outReportPath")

	// Ninja rule to execute go build.
	goBuild = pctx.StaticRule("binaryBuild", blueprint.RuleParams{
		Command:     "cd $workDir && go build -o $outputPath $pkg",
		Description: "build go command $pkg",
	}, "workDir", "outputPath", "pkg")

	// Ninja rule to execute go mod vendor.
	goVendor = pctx.StaticRule("vendor", blueprint.RuleParams{
		Command:     "cd $workDir && go mod vendor",
		Description: "vendor dependencies of $name",
	}, "workDir", "name")
)

// testedBinaryModule implements the simplest Go binary build with running tests for the target Go package.
type testedBinaryModule struct {
	blueprint.SimpleName

	properties struct {
		// Go package name to build as a command with "go build".
		Pkg string
		// Go package name to test as a command with "go test"
		TestPkg string
		// List of source files.
		Srcs []string
		// Exclude patterns.
		SrcsExclude []string
		// If to call vendor command.
		VendorFirst bool

		// Example of how to specify dependencies.
		Deps []string
	}
}

func sliceIncludes(element string, slice []string) bool {
	includes := false

	for _, v := range slice {
		if element == v {
			includes = true
		}
	}

	return includes
}

func (gb *testedBinaryModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	testReportName := name + ".txt"
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding build actions for go binary module '%s'", name)

	outputPath := path.Join(config.BaseOutputDir, "bin", name)
	outReportPath := path.Join(config.BaseOutputDir, "reports", testReportName)

	var inputs []string // files which will be passed to "go build"
	var testInputs []string // files which will be passed to "go test", includes all golang files to watch to changes not only test files
	inputErrors := false
	for _, src := range gb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, gb.properties.SrcsExclude); err == nil {
			testInputs = append(testInputs, matches...)

			for _, input := range matches {
				if !strings.Contains(input, "_test.go") && !sliceIncludes(input, inputs) {
					inputs = append(inputs, input)
				}
			}
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErrors = true
		}
	}
	if inputErrors {
		return
	}

	if gb.properties.VendorFirst {
		vendorDirPath := path.Join(ctx.ModuleDir(), "vendor")
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Vendor dependencies of %s", name),
			Rule:        goVendor,
			Outputs:     []string{vendorDirPath},
			Implicits:   []string{path.Join(ctx.ModuleDir(), "go.mod")},
			Optional:    true,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"name":    name,
			},
		})
		inputs = append(inputs, vendorDirPath)
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go binary", name),
		Rule:        goBuild,
		Outputs:     []string{outputPath},
		Implicits:   inputs,
		Args: map[string]string{
			"outputPath": outputPath,
			"workDir":    ctx.ModuleDir(),
			"pkg":        gb.properties.Pkg,
		},
	})

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go test report", testReportName),
		Rule:        goTest,
		Outputs:     []string{outReportPath},
		Implicits:   testInputs,
		Args: map[string]string{
			"outReportPath": outReportPath,
			"workDir":              ctx.ModuleDir(),
			"testPkg":              gb.properties.TestPkg,
		},
	})

}

// SimpleBinFactory is a factory for go binary module type which supports Go command packages with running tests.
func SimpleBinFactory() (blueprint.Module, []interface{}) {
	mType := &testedBinaryModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
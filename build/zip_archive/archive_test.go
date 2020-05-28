package zip_archive

import (
	"bytes"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"strings"
	"testing"
)

func TestSimpleArchiveFactory(t *testing.T) {
	ctx := blueprint.NewContext()

	ctx.MockFileSystem(map[string][]byte{
		"Blueprints": []byte(`
		 zip_archive{
		  name: "test-archive",
          srcs: ["**/*.txt"],
		 }
		`),
		"test3.txt":              nil,
		"test2/test2.txt":        nil,
		"test1/test11/test1.txt": nil,
	})

	ctx.RegisterModuleType("zip_archive", SimpleArchiveFactory)

	cfg := bood.NewConfig()

	_, errs := ctx.ParseBlueprintsFiles(".", cfg)
	if len(errs) != 0 {
		t.Fatalf("Syntax errors in the test blueprint file: %s", errs)
	}

	_, errs = ctx.PrepareBuildActions(cfg)
	if len(errs) != 0 {
		t.Errorf("Unexpected errors while preparing build actions: %s", errs)
	}
	buffer := new(bytes.Buffer)
	if err := ctx.WriteBuildFile(buffer); err != nil {
		t.Errorf("Error writing ninja file: %s", err)
	} else {
		text := buffer.String()
		t.Logf("Gennerated ninja build file:\n%s", text)
		if !strings.Contains(text, "command = cd ${workDir} && zip ${outputPath} -j ${inputFiles}") {
			t.Errorf("Generated ninja file have uncorrect makeArchive rule")
		}
		if !strings.Contains(text, "build out/archives/test-archive.zip: ") {
			t.Errorf("Generated ninja file does not have(or have uncorrect) build of the archive module")
		}
		if !strings.Contains(text, "inputFiles = test3.txt,test1/test11/test1.txt,test2/test2.txt") {
			t.Errorf("Generated ninja file does not have correct list of input files")
		}
	}
}

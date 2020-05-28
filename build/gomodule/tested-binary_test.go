package gomodule

import (
	"bytes"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"strings"
	"testing"
)

func TestSimpleBinFactory(t *testing.T) {
	ctx := blueprint.NewContext()

	ctx.MockFileSystem(map[string][]byte{
		"Blueprints": []byte(`
			go_binary {
			  name: "test-out",
			  srcs: ["test-src.go", "test-src_test.go"],
			  pkg: ".",
			  testPkg: ".",
	          vendorFirst: true
			}
		`),
		"test-src.go":      nil,
		"test-src_test.go": nil,
	})

	ctx.RegisterModuleType("go_binary", SimpleBinFactory)

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
		if !strings.Contains(text, "out/bin/test-out: ") {
			t.Errorf("Generated ninja file does not have build of the test module")
		}
		if !strings.Contains(text, " test-src.go") {
			t.Errorf("Generated ninja file does not have source dependency")
		}
		if !strings.Contains(text, "build vendor: g.gomodule.vendor | go.mod") {
			t.Errorf("Generated ninja file does not have vendor build rule")
		}
		if !strings.Contains(text, "command = cd ${workDir} && go test -v ${testPkg} > ${outReportPath}") {
			t.Errorf("Generated ninja file does have uncorrect testing rule")
		}
		if !strings.Contains(text, "build out/reports/test-out.txt: ") {
			t.Errorf("Generated ninja file does not have tests build rule")
		}
		if !strings.Contains(text, "test-src_test.go") {
			t.Errorf("Generated ninja file does not have testing file")
		}
	}
}

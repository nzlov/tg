package generate

import (
	"fmt"
	"go/ast"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.Buf, format, args...)
}

func (g *Generator) Format() []byte {
	src, err := format.Source(g.Buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.Buf.Bytes()
	}
	return src
}

func (g *Generator) ParsePackage(patterns []string, tags []string) {
	cfg := &packages.Config{
		Mode:       packages.LoadSyntax,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.AddPackage(pkgs[0])
}

func (g *Generator) AddPackage(pkg *packages.Package) {
	g.Pkg = &Package{
		Name:  pkg.Name,
		Path:  pkg.PkgPath,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}
	g.Project = strings.Join(strings.Split(pkg.PkgPath, "/")[:3], "/")

	for i, file := range pkg.Syntax {
		g.Pkg.files[i] = &File{
			g:           g,
			file:        file,
			imp:         make([]string, 0),
			pkg:         g.Pkg,
			mappers:     []Mapper{},
			trimPrefix:  g.TrimPrefix,
			lineComment: g.LineComment,
		}
	}
}

func (g *Generator) Generate() {

	mappers := make([]Mapper, 0, 100)
	for _, file := range g.Pkg.files {
		file.mappers = nil
		if file.file != nil {
			ast.Inspect(file.file, file.genDecl)
			mappers = append(mappers, file.mappers...)
		}
	}

	for _, m := range mappers {
		r := m.Render()
		//		data, _ := json.MarshalIndent(&r, "", "  ")
		//		log.Println("R:", string(data))
		{
			// init.go
			g.Buf.Reset()

			err := initT.Execute(&g.Buf, &r)
			if err != nil {
				panic(err)
			}

			src := g.Format()
			path := filepath.Join(g.Output, r.PackageName)
			os.MkdirAll(path, os.ModePerm)
			outputName := filepath.Join(path, "i.go")
			err = ioutil.WriteFile(outputName, src, 0644)
			if err != nil {
				log.Fatalf("writing output: %s", err)
			}
		}
		{
			g.Buf.Reset()

			err := cT.Execute(&g.Buf, &r)
			if err != nil {
				panic(err)
			}

			src := g.Format()
			path := filepath.Join(g.Output, r.PackageName)
			os.MkdirAll(path, os.ModePerm)
			outputName := filepath.Join(path, "c.go")
			err = ioutil.WriteFile(outputName, src, 0644)
			if err != nil {
				log.Fatalf("writing output: %s", err)
			}
		}
	}

}

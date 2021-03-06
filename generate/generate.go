package generate

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"log"
	"math"
	"os"
	exec "os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"text/template"

	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/packages"
)

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
		logrus.Fatalf("error: %d packages found", len(pkgs))
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

	if g.Template != "" {
		t, err := template.ParseFiles(g.Template)
		if err != nil {
			logrus.Fatalln("Load custom template[", g.Template, "] error:", err)
		}
		cT = t
	}

	mappers := make([]Mapper, 0, 100)
	for _, file := range g.Pkg.files {
		file.mappers = nil
		if file.file != nil {
			ast.Inspect(file.file, file.genDecl)
			if g.Debug {
				fset := token.NewFileSet()
				ast.Print(fset, file.file)
			}

			mappers = append(mappers, file.mappers...)
		}
	}

	if len(mappers) < g.gonum {
		g.gonum = int(math.Log(float64(g.gonum)))
	}

	mChan := make(chan Mapper, g.gonum)

	w := &sync.WaitGroup{}
	for i := 0; i < g.gonum; i++ {
		w.Add(1)
		go func(i int) {
			buf := bytes.NewBufferString("")
			defer func() {
				w.Done()
				logrus.Infoln("G", i, "Done")
				if err := recover(); err != nil {
					debug.PrintStack()
					log.Fatalln(err)
				}
			}()
			for {
				v, ok := <-mChan
				if !ok {
					break
				}
				logrus.Infoln("G", i, v.Name)
				r := v.Render()
				//		data, _ := json.MarshalIndent(&r, "", "  ")
				//		log.Println("R:", string(data))
				{
					buf.Reset()
					err := cT.Execute(buf, &r)
					if err != nil {
						panic(err)
					}

					path := filepath.Join(g.Output, r.PackageName)
					os.MkdirAll(path, os.ModePerm)
					outputName := filepath.Join(path, "tg.go")
					err = ioutil.WriteFile(outputName, buf.Bytes(), 0644)
					if err != nil {
						logrus.Fatalf("writing output: %s", err)
					}
				}
			}

		}(i)
	}

	for _, m := range mappers {
		mChan <- m
	}
	close(mChan)
	w.Wait()

	logrus.Infoln("Gen Done")
	logrus.Infoln("Start Format")

	_, err := exec.Command("goimports", "-w", g.Output).Output()
	if err != nil {
		logrus.Fatalln("Auto Format error:", err.Error())
	}

	logrus.Infoln("Done")
}

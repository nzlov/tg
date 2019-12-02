package generate

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"strings"
)

const (
	FuncType_CreateBefore   = "CreateBefore"
	FuncType_CreateTxBefore = "CreateTxBefore"
	FuncType_CreateTxAfter  = "CreateTxAfter"
	FuncType_CreateAfter    = "CreateAfter"
	FuncType_UpdateBefore   = "UpdateBefore"
	FuncType_UpdateTxBefore = "UpdateTxBefore"
	FuncType_UpdateTxAfter  = "UpdateTxAfter"
	FuncType_UpdateAfter    = "UpdateAfter"
	FuncType_InfoBefore     = "InfoBefore"
	FuncType_InfoAfter      = "InfoAfter"
	FuncType_ListBefore     = "ListBefore"
	FuncType_ListAfter      = "ListAfter"
	FuncType_DeleteBefore   = "DeleteBefore"
	FuncType_DeleteTxBefore = "DeleteTxBefore"
	FuncType_DeleteTxAfter  = "DeleteTxAfter"
	FuncType_DeleteAfter    = "DeleteAfter"
)

type Generator struct {
	Buf     bytes.Buffer
	Pkg     *Package
	Project string

	Output string

	TrimPrefix  string
	LineComment bool

	Func map[string][]Func // 所有的ModelController都需要的方法
}

func NewGenerator(trimprefix, output string, linecomment bool) *Generator {
	return &Generator{
		TrimPrefix:  trimprefix,
		LineComment: linecomment,
		Output:      output,

		Func: map[string][]Func{},
	}
}

// Value represents a declared constant.
type Mapper struct {
	File    *File
	Name    string
	Attr    []Attr
	DBIndex string
	API     string
}

func (m Mapper) Render() Render {

	r := Render{
		Args:        strings.Join(os.Args[1:], " "),
		PackageName: strings.ToLower(m.Name) + "s",
		Project:     m.File.g.Project,
		Name:        m.Name,
		DBIndex:     m.DBIndex,
		CreateSave:  true,
		UpdateSave:  true,
	}

	for k, v := range m.File.g.Func {
		mf := []Func{}
		for _, f := range v {
			if f.Check(m.Name) {
				mf = append(mf, f)
			}
		}
		switch k {
		case FuncType_CreateBefore:
			r.CreateBefore = mf
		case FuncType_CreateTxBefore:
			r.CreateTxBefore = mf
		case FuncType_CreateTxAfter:
			r.CreateTxAfter = mf
		case FuncType_CreateAfter:
			r.CreateAfter = mf
		case FuncType_UpdateBefore:
			r.UpdateBefore = mf
		case FuncType_UpdateTxBefore:
			r.UpdateTxBefore = mf
		case FuncType_UpdateTxAfter:
			r.UpdateTxAfter = mf
		case FuncType_UpdateAfter:
			r.UpdateAfter = mf
		case FuncType_InfoBefore:
			r.InfoBefore = mf
		case FuncType_InfoAfter:
			r.InfoAfter = mf
		case FuncType_ListBefore:
			r.ListBefore = mf
		case FuncType_ListAfter:
			r.ListAfter = mf
		case FuncType_DeleteBefore:
			r.DeleteBefore = mf
		case FuncType_DeleteTxBefore:
			r.DeleteTxBefore = mf
		case FuncType_DeleteTxAfter:
			r.DeleteTxAfter = mf
		case FuncType_DeleteAfter:
			r.DeleteAfter = mf

		}

	}

	{
		v := strings.Split(m.API, " ")
		if len(v) == 1 {
			r.Create = true
			r.Update = true
			r.List = true
			r.Info = true
			r.Delete = true
		} else {
			r.Create = true
			r.Update = true
			r.List = true
			r.Info = true
			r.Delete = true

			for _, vv := range v {

				if strings.HasPrefix(vv, "-") {
					switch vv {
					case "Create":
						r.Create = false
					case "Update":
						r.Update = false
					case "List":
						r.List = false
					case "Info":
						r.Info = false
					case "Delete":
						r.Delete = false
					}
					continue
				}

				vvs := strings.Split(vv, ":")

				if len(vvs) == 2 {
					ops := strings.Split(vvs[1], ";")
					for _, v := range ops {
						vs := strings.Split(v, "=")
						switch vs[0] {
						case "nosave":
							switch vvs[0] {
							case "Create":
								r.CreateSave = false
							case "Update":
								r.UpdateSave = false
							}
						case "preload":
							pvs := strings.Split(strings.ReplaceAll(vs[1], ">", "\":\""), ",")
							switch vvs[0] {
							case "Info":
								r.InfoPreload = true
								r.InfoPreloadV = pvs
							case "List":
								r.ListPreload = true
								r.ListPreloadV = pvs
							}
						}
					}
				}
			}

		}
	}
	if r.Create {
		r.CreateParams = []Attr{}
		r.CreateParamsDecs = []string{}
		for _, v := range m.Attr {
			if p := v.Param("create"); p != "" {
				r.CreateParamsDecs = append(r.CreateParamsDecs, p)
				r.CreateParams = append(r.CreateParams, v)
			}
		}
	}
	if r.Update {
		r.UpdateParams = []Attr{}
		r.UpdateParamsDecs = []string{}
		for _, v := range m.Attr {
			if p := v.Param("update"); p != "" {
				r.UpdateParamsDecs = append(r.UpdateParamsDecs, p)
				r.UpdateParams = append(r.UpdateParams, v)
			}
		}
	}

	return r
}

type Attr struct {
	Name      string
	Type      string
	CtxFunc   string
	JSON      string
	Enums     string
	MaxLength string
	MinLength string
	Max       string
	Min       string
	Params    string
	Desc      string
}

// Param 获取接口文档参数说明
func (a Attr) Param(t string) string {
	h := false
	require := false
	switch t {
	case "create":
		switch a.Params {
		case "CU", "Cu", "C":
			require = true
			fallthrough
		case "cU", "cu", "c":
			h = true
		}
	case "update":
		switch a.Params {
		case "CU", "cU", "U":
			require = true
			fallthrough
		case "Cu", "cu", "u":
			h = true
		}
	}
	if h {
		f := "// @Param %s formData %s %v \"%s\""
		as := []interface{}{a.JSON, a.Type, require, a.Desc}
		if a.Enums != "" {
			f += " %s"
			as = append(as, "enums("+a.Enums+")")
		}
		if a.MaxLength != "" {
			f += " %s"
			as = append(as, "maxLength("+a.MaxLength+")")
		}
		if a.MinLength != "" {
			f += " %s"
			as = append(as, "minLength("+a.MinLength+")")
		}
		if a.Max != "" {
			f += " %s"
			as = append(as, "maxinum("+a.Max+")")
		}
		if a.Min != "" {
			f += " %s"
			as = append(as, "mininum("+a.Min+")")
		}
		return fmt.Sprintf(f, as...)
	}
	return ""
}

type Func struct {
	Name     string
	Includes map[string]struct{}
	Excludes map[string]struct{}
}

func (f *Func) Check(n string) bool {
	if f.Includes != nil {
		_, ok := f.Includes[n]
		return ok
	}
	if f.Excludes != nil {
		_, ok := f.Excludes[n]
		return !ok
	}
	return true
}

type File struct {
	g *Generator

	imp  []string
	pkg  *Package
	file *ast.File

	mappers []Mapper

	trimPrefix  string
	lineComment bool
}

type Package struct {
	Name  string
	Path  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

type Render struct {
	Args        string
	PackageName string
	Project     string
	Name        string
	DBIndex     string

	Create           bool
	CreateSave       bool
	CreateParams     []Attr
	CreateParamsDecs []string
	CreateBefore     []Func
	CreateTxBefore   []Func
	CreateTxAfter    []Func
	CreateAfter      []Func

	Update           bool
	UpdateSave       bool
	UpdateParams     []Attr
	UpdateParamsDecs []string
	UpdateBefore     []Func
	UpdateTxBefore   []Func
	UpdateTxAfter    []Func
	UpdateAfter      []Func

	List         bool
	ListPreload  bool
	ListPreloadV []string
	ListBefore   []Func
	ListAfter    []Func

	Info         bool
	InfoPreload  bool
	InfoPreloadV []string
	InfoBefore   []Func
	InfoAfter    []Func

	Delete         bool
	DeleteBefore   []Func
	DeleteTxBefore []Func
	DeleteTxAfter  []Func
	DeleteAfter    []Func
}

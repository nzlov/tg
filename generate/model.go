package generate

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"sort"
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
	gonum   int
	Pkg     *Package
	Project string

	Output string
	Debug  bool

	TrimPrefix  string
	LineComment bool
	Template    string

	Func map[string][]Func // 所有的ModelController都需要的方法
}

func NewGenerator(gonum int, trimprefix, output string, linecomment, debug bool, template string) *Generator {
	return &Generator{
		gonum:       gonum,
		TrimPrefix:  trimprefix,
		LineComment: linecomment,
		Output:      output,
		Debug:       debug,
		Template:    template,

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
		Desc:        m.Name,
	}

	for k, v := range m.File.g.Func {
		mfs := []MFunc{}
		for _, f := range v {
			if mf := f.Check(m.Name); mf != nil {
				mfs = append(mfs, *mf)
			}
		}

		sort.Sort(MFuncs(mfs))

		switch k {
		case FuncType_CreateBefore:
			r.CreateBefore = mfs
		case FuncType_CreateTxBefore:
			r.CreateTxBefore = mfs
		case FuncType_CreateTxAfter:
			r.CreateTxAfter = mfs
		case FuncType_CreateAfter:
			r.CreateAfter = mfs
		case FuncType_UpdateBefore:
			r.UpdateBefore = mfs
		case FuncType_UpdateTxBefore:
			r.UpdateTxBefore = mfs
		case FuncType_UpdateTxAfter:
			r.UpdateTxAfter = mfs
		case FuncType_UpdateAfter:
			r.UpdateAfter = mfs
		case FuncType_InfoBefore:
			r.InfoBefore = mfs
		case FuncType_InfoAfter:
			r.InfoAfter = mfs
		case FuncType_ListBefore:
			r.ListBefore = mfs
		case FuncType_ListAfter:
			r.ListAfter = mfs
		case FuncType_DeleteBefore:
			r.DeleteBefore = mfs
		case FuncType_DeleteTxBefore:
			r.DeleteTxBefore = mfs
		case FuncType_DeleteTxAfter:
			r.DeleteTxAfter = mfs
		case FuncType_DeleteAfter:
			r.DeleteAfter = mfs

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

				if vv == "nosave" {
					r.CreateSave = false
					r.UpdateSave = false
				}
				if strings.HasPrefix(vv, "desc") {
					vs := strings.Split(vv, "=")
					if len(vs) > 1 {
						r.Desc = vs[1]
					}
				}

				if strings.HasPrefix(vv, "preload") {
					vs := strings.Split(vv, "=")
					pvs := strings.Split(strings.ReplaceAll(vs[1], ">", "\":\""), ",")
					r.InfoPreload = true
					r.InfoPreloadV = pvs
					r.ListPreload = true
					r.ListPreloadV = pvs
				}

				if strings.HasPrefix(vv, "security") {
					vs := strings.Split(vv, "=")
					sec := strings.Split(vs[1], ",")
					r.CreateSecurity = sec
					r.UpdateSecurity = sec
					r.ListSecurity = sec
					r.InfoSecurity = sec
					r.DeleteSecurity = sec

				}

				if strings.HasPrefix(vv, "-") {
					switch vv {
					case "-Create":
						r.Create = false
					case "-Update":
						r.Update = false
					case "-List":
						r.List = false
					case "-Info":
						r.Info = false
					case "-Delete":
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
						case "save":
							switch vvs[0] {
							case "Create":
								r.CreateSave = true
							case "Update":
								r.UpdateSave = true
							}
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
						case "security":
							sec := strings.Split(vs[1], ",")
							switch vvs[0] {
							case "Create":
								r.CreateSecurity = sec
							case "Update":
								r.UpdateSecurity = sec
							case "Info":
								r.InfoSecurity = sec
							case "List":
								r.ListSecurity = sec
							case "Delete":
								r.DeleteSecurity = sec
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
	IToM      string
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

type MFunc struct {
	Name string
	Sort int64
}

type MFuncs []MFunc

func (m MFuncs) Len() int {
	return len(m)
}
func (m MFuncs) Less(i, j int) bool {
	return m[i].Sort > m[j].Sort
}
func (m MFuncs) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type Func struct {
	Name     string
	Sort     int64
	Includes map[string]int64
	Excludes map[string]struct{}
}

func (f Func) Check(n string) *MFunc {
	if f.Includes != nil {
		if v, ok := f.Includes[n]; ok {
			if v == -1 {
				v = f.Sort
			}
			return &MFunc{Name: f.Name, Sort: v}
		}
		return nil
	}
	if f.Excludes != nil {
		_, ok := f.Excludes[n]
		if ok {
			return nil
		}
	}
	return &MFunc{Name: f.Name, Sort: f.Sort}
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
	Desc        string

	Create           bool
	CreateSave       bool
	CreateParams     []Attr
	CreateParamsDecs []string
	CreateBefore     []MFunc
	CreateTxBefore   []MFunc
	CreateTxAfter    []MFunc
	CreateAfter      []MFunc
	CreateSecurity   []string

	Update           bool
	UpdateSave       bool
	UpdateParams     []Attr
	UpdateParamsDecs []string
	UpdateBefore     []MFunc
	UpdateTxBefore   []MFunc
	UpdateTxAfter    []MFunc
	UpdateAfter      []MFunc
	UpdateSecurity   []string

	List         bool
	ListPreload  bool
	ListPreloadV []string
	ListBefore   []MFunc
	ListAfter    []MFunc
	ListSecurity []string

	Info         bool
	InfoPreload  bool
	InfoPreloadV []string
	InfoBefore   []MFunc
	InfoAfter    []MFunc
	InfoSecurity []string

	Delete         bool
	DeleteBefore   []MFunc
	DeleteTxBefore []MFunc
	DeleteTxAfter  []MFunc
	DeleteAfter    []MFunc
	DeleteSecurity []string
}

package generate

import (
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

func (f *File) genDecl(node ast.Node) bool {
	switch t := node.(type) {
	case *ast.GenDecl:
		switch t.Tok {
		case token.IMPORT:
			for _, spec := range t.Specs {
				ispec := spec.(*ast.ImportSpec)
				f.imp = append(f.imp, ispec.Path.Value)
			}
		case token.TYPE:
			if t.Doc != nil {
				if strings.HasPrefix(t.Doc.Text(), "@tg") {
					m := Mapper{
						File: f,
						API:  strings.TrimSpace(t.Doc.Text()),
						Attr: []Attr{},
					}
					for _, spec := range t.Specs {
						switch st := spec.(type) {
						case *ast.TypeSpec:
							m.Name = st.Name.String()
							switch stf := st.Type.(type) {
							case *ast.StructType:
								for _, field := range stf.Fields.List {
									switch ft := field.Type.(type) {
									case *ast.Ident:
										if field.Tag == nil {
											continue
										}
										structTag := reflect.StructTag(strings.Replace(field.Tag.Value, "`", "", -1))

										if v, ok := structTag.Lookup("dbindex"); ok {
											m.DBIndex = v
										}

										at := Attr{}
										at.Params = structTag.Get("params")
										if at.Params == "" {
											continue
										}

										at.JSON = structTag.Get("json")
										at.Enums = structTag.Get("enums")
										at.MaxLength = structTag.Get("maxlength")
										at.MinLength = structTag.Get("minlength")
										at.Max = structTag.Get("max")
										at.Min = structTag.Get("min")

										switch ft.Name {
										case "bool":
											at.Type = "bool"
											at.CtxFunc = "Bool"
										case "int", "int8", "int16", "int32", "int64":
											at.Type = "integer"
											at.CtxFunc = "Int64"
										case "float32", "float64":
											at.Type = "number"
											at.CtxFunc = "Float64"
										default:
											at.Type = "string"
											at.CtxFunc = "String"
										}
										at.Name = strings.TrimSpace(field.Names[0].Name)
										if field.Doc != nil {
											at.Desc = strings.TrimSpace(field.Doc.Text())
										} else {
											at.Desc = strings.TrimSpace(at.Name)
										}
										m.Attr = append(m.Attr, at)
									}
								}
							}
						}
					}
					f.mappers = append(f.mappers, m)
				}
			}
		}
		return false
	case *ast.FuncDecl:
		if t.Doc != nil {
			if strings.HasPrefix(t.Doc.Text(), "@tg") {

				tgs := strings.Split(t.Doc.Text()[4:], " ")
				for _, v := range tgs {
					fc := Func{
						Name: t.Name.String(),
					}
					v = strings.TrimSpace(v)
					if strings.Index(v, ":") > -1 {
						vs := strings.Split(v, ":")
						vsv := strings.Split(vs[1], ",")

						if strings.HasPrefix(v, "-") {
							m := map[string]struct{}{}
							for _, s := range vsv {
								m[strings.TrimSpace(s)] = struct{}{}
							}
							fc.Excludes = m
						} else {
							m := map[string]int64{}
							for _, s := range vsv {
								if i := strings.LastIndex(s, "@"); i > -1 {
									// 存在排序
									vs := strings.Split(s, "@")
									m[strings.TrimSpace(s[:i])], _ = strconv.ParseInt(vs[1], 10, 64)
								} else {
									m[strings.TrimSpace(s)] = -1
								}
							}
							fc.Includes = m
						}
						v = strings.Trim(vs[0], "-")
					} else {
						if i := strings.LastIndex(v, "@"); i > -1 {
							// 存在排序
							vs := strings.Split(v, "@")
							fc.Sort, _ = strconv.ParseInt(vs[1], 10, 64)
							v = v[:i]
						}
					}
					if fs, ok := f.g.Func[v]; ok {
						fs = append(fs, fc)
						f.g.Func[v] = fs
					} else {
						f.g.Func[v] = []Func{fc}
					}
				}
			}
		}

		return false
	}

	return true
}

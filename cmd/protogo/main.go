package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
)

var path = flag.String("path", "", "path to go package")

type visitor struct {
	exportedTypes map[*ast.TypeSpec]struct{}
	fieldTags     map[*ast.StructType]map[int]struct{}
}

var fieldTagRegexp = regexp.MustCompile(`protogo:"(\d+)"`)

func (v *visitor) exportStruct(s *ast.StructType) {
	v.fieldTags[s] = make(map[int]struct{})
	for _, f := range s.Fields.List {
		if !f.Names[0].IsExported() {
			continue
		}
		if f.Tag == nil {
			log.Fatalf("expected 'protogo' field tag")
		}
		tag := f.Tag.Value // check if nil
		matches := fieldTagRegexp.FindStringSubmatch(tag)
		if len(matches) != 2 {
			log.Fatalf("expected one 'protogo' field tag, got %d", len(matches)-1) // first match is of whole expression
		}
		tagNum, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Fatalf("Bad protogo field tag number: '%s'", tag)
		}
		log.Printf("Tag: %d", tagNum)
		if _, ok := v.fieldTags[s][tagNum]; ok {
			log.Fatalf("Duplicate field tag number %d", tagNum)
		}
		v.fieldTags[s][tagNum] = struct{}{}
	}
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	switch x := n.(type) {
	case *ast.GenDecl:
		for _, spec := range x.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				if !s.Name.IsExported() {
					break
				}
				v.exportedTypes[s] = struct{}{}
				log.Printf("Declaring: %s", s.Name.Name)
				switch t := s.Type.(type) {
				case *ast.StructType:
					v.exportStruct(t)
				}
			case *ast.ValueSpec:
				for i := 0; i < len(s.Names); i++ {
					if !s.Names[i].IsExported() {
						continue
					}

					if _, isStruct := s.Type.(*ast.StructType); isStruct {
						log.Fatalf("Cannot use anonymous struct for exported variable (%s)", s.Names[i])
					}
					// TODO: auto-export anonymous structs of globals?
					// v.exportStruct(st)

					if _, ok := v.exportedTypes[s.Type.(*ast.Ident).Obj.Decl.(*ast.TypeSpec)]; !ok {
						log.Fatalf("Struct type must be separately declared")
					}
				}
			default:
				log.Println("Skipping:", reflect.TypeOf(s))
			}
		}
	case *ast.FuncDecl:
		for i, p := range x.Type.Params.List {
			if p.Type.(*ast.Ident).Obj == nil {
				log.Fatalf("Exported function must use protogo struct (%s, argument %d)", x.Name.Name, i)
			}
			_, ok := v.exportedTypes[p.Type.(*ast.Ident).Obj.Decl.(*ast.TypeSpec)]
			if !ok {
				log.Fatalf("Exported function must use protogo struct (%s, argument %d)", x.Name.Name, i)
			}
		}
		for i, r := range x.Type.Results.List {
			if r.Type.(*ast.Ident).Obj == nil {
				log.Fatalf("Exported function must use protogo struct (%s, result %d)", x.Name.Name, i)
			}
			_, ok := v.exportedTypes[r.Type.(*ast.Ident).Obj.Decl.(*ast.TypeSpec)]
			if !ok {
				log.Fatalf("Exported function must use protogo struct (%s, result %d)", x.Name.Name, i)
			}
		}
	default:
		log.Println("Skipping:", reflect.TypeOf(x))
	}
	return v
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if *path == "" {
		log.Fatalln("Must specify path")
	}

	fset := token.NewFileSet()
	asts, err := parser.ParseDir(fset, *path, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	if len(asts) != 1 {
		log.Fatalf("path must contain exactly one package, found %d", len(asts))
	}
	packageName := ""
	for p := range asts {
		// should happen only once
		packageName = p
	}
	log.Println("Protogo package:", packageName)
	pkg := ast.MergePackageFiles(asts[packageName], ast.FilterImportDuplicates)
	ast.Fprint(os.Stdout, fset, pkg, nil)
	v := &visitor{
		exportedTypes: make(map[*ast.TypeSpec]struct{}),
		fieldTags:     make(map[*ast.StructType]map[int]struct{}),
	}
	ast.Walk(v, pkg)

	// TODO: split flow
	//        - walk tree adding types to shouldBeExported map[ast.TypeSpec]
	//          (types used in exported vars, transitive types, etc.)
	//        - validate each type in shouldBeExported (struct, unique protogo ids)
	//        - synthesize protocol buffer for each type in shouldBeExported
	//        - rewrite declarations from shouldBeExported type to protocol buffer
	//        - rewrite accesses to each var to the correct protocol buffer equivalent

	// TODO: from v.exportedTypes, synthesize the protocol buffers (and gRPC?)

	// TODO: synthesize main
}

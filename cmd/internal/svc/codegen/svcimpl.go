package codegen

import (
	"bytes"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"github.com/unionj-cloud/go-doudou/v2/toolkit/astutils"
	"github.com/unionj-cloud/go-doudou/v2/toolkit/copier"
	v3 "github.com/unionj-cloud/go-doudou/v2/toolkit/protobuf/v3"
	"github.com/unionj-cloud/go-doudou/v2/version"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

var svcimportTmpl = `
	"context"
	"{{.ConfigPackage}}"
	"{{.DtoPackage}}"
	"github.com/jmoiron/sqlx"
	"github.com/brianvoe/gofakeit/v6"
`

var appendPart = `{{- range $m := .Meta.Methods }}
	func (receiver *{{$.Meta.Name}}Impl) {{$m.Name}}({{- range $i, $p := $m.Params}}
    {{- if $i}},{{end}}
    {{- $p.Name}} {{$p.Type}}
    {{- end }}) ({{- range $i, $r := $m.Results}}
                     {{- if $i}},{{end}}
                     {{- $r.Name}} {{$r.Type}}
                     {{- end }}) {
    	var _result struct{
			{{- range $r := $m.Results }}
			{{- if ne $r.Type "error" }}
			{{ $r.Name | toCamel }} {{ $r.Type }}
			{{- end }}
			{{- end }}
		}
		_ = gofakeit.Struct(&_result)
		return {{range $i, $r := $m.Results }}{{- if $i}},{{end}}{{ if eq $r.Type "error" }}nil{{else}}_result.{{ $r.Name | toCamel }}{{end}}{{- end }}
    }
{{- end }}`

var svcimplTmpl = `/**
* Generated by go-doudou {{.Version}}.
* You can edit it as your need.
*/
package {{.SvcPackage}}

import ()

var _ {{.Meta.Name}} = (*{{.Meta.Name}}Impl)(nil)

type {{.Meta.Name}}Impl struct {
	conf *config.Config
}

` + appendPart + `

func New{{.Meta.Name}}(conf *config.Config) *{{.Meta.Name}}Impl {
	return &{{.Meta.Name}}Impl{
		conf: conf,
	}
}
`

var svcimportTmplGrpc = `
	"context"
	"{{.ConfigPackage}}"
	"{{.DtoPackage}}"
	pb "{{.PbPackage}}"
`

var appendPartGrpc = `{{- range $m := .GrpcSvc.Rpcs }}
    {{- if eq $m.StreamType 0 }}
	func (receiver *{{$.Meta.Name}}Impl) {{$m.Name}}(ctx context.Context, request *{{$m.Request | convert}}) (*{{$m.Response | convert}}, error) {
    	//TODO implement me
		panic("implement me")
    }
    {{- end }}
    {{- if eq $m.StreamType 1 }}
	func (receiver *{{$.Meta.Name}}Impl) {{$m.Name}}(server pb.{{$.GrpcSvc.Name}}_{{$m.Name}}Server) error {
		//TODO implement me
		panic("implement me")
	}
    {{- end }}
    {{- if eq $m.StreamType 2 }}
	func (receiver *{{$.Meta.Name}}Impl) {{$m.Name}}(server pb.{{$.GrpcSvc.Name}}_{{$m.Name}}Server) error {
		//TODO implement me
		panic("implement me")
	}
    {{- end }}
    {{- if eq $m.StreamType 3 }}
	func (receiver *{{$.Meta.Name}}Impl) {{$m.Name}}(request *{{$m.Request | convert}}, server pb.{{$.GrpcSvc.Name}}_{{$m.Name}}Server) error {
		//TODO implement me
		panic("implement me")
	}
    {{- end }}
{{- end }}`

/**
type Service struct {
	Name      string
	Package   string
	GoPackage string
	Syntax    string
	// go-doudou version
	Version  string
	ProtoVer string
	Rpcs     []Rpc
	Messages []Message
	Enums    []Enum
	Comments []string
	Imports  []string
}
*/
var svcimplTmplGrpc = `/**
* Generated by go-doudou {{.Version}}.
* You can edit it as your need.
*/
package {{.SvcPackage}}

import ()

var _ pb.{{.GrpcSvc.Name}}Server = (*{{.Meta.Name}}Impl)(nil)

type {{.Meta.Name}}Impl struct {
    pb.Unimplemented{{.GrpcSvc.Name}}Server
	conf *config.Config
}

` + appendPartGrpc + `

func New{{.Meta.Name}}(conf *config.Config) *{{.Meta.Name}}Impl {
	return &{{.Meta.Name}}Impl{
		conf: conf,
	}
}
`

// GenSvcImpl generates service implementation
func GenSvcImpl(dir string, ic astutils.InterfaceCollector) {
	var (
		err         error
		svcimplfile string
		f           *os.File
		tpl         *template.Template
		buf         bytes.Buffer
		meta        astutils.InterfaceMeta
		tmpl        string
		importBuf   bytes.Buffer
	)
	svcimplfile = filepath.Join(dir, "svcimpl.go")
	err = copier.DeepCopy(ic.Interfaces[0], &meta)
	if err != nil {
		panic(err)
	}
	cfgPkg := astutils.GetPkgPath(filepath.Join(dir, "config"))
	dtoPkg := astutils.GetPkgPath(filepath.Join(dir, "dto"))
	if _, err = os.Stat(svcimplfile); os.IsNotExist(err) {
		if f, err = os.Create(svcimplfile); err != nil {
			panic(err)
		}
		defer f.Close()
		tmpl = svcimplTmpl
	} else {
		logrus.Warningln("New content will be append to file svcimpl.go")
		if f, err = os.OpenFile(svcimplfile, os.O_APPEND, os.ModePerm); err != nil {
			panic(err)
		}
		defer f.Close()
		tmpl = appendPart

		sc := astutils.NewStructCollector(astutils.ExprString)
		astutils.CollectStructsInFolder(dir, sc)
		if implementations, exists := sc.Methods[meta.Name+"Impl"]; exists {
			var notimplemented []astutils.MethodMeta
			for _, item := range meta.Methods {
				for _, implemented := range implementations {
					if item.Name == implemented.Name {
						goto L
					}
				}
				notimplemented = append(notimplemented, item)

			L:
			}

			meta.Methods = notimplemented
		}
	}

	funcMap := make(map[string]interface{})
	funcMap["toCamel"] = strcase.ToCamel
	if tpl, err = template.New("svcimpl.go.tmpl").Funcs(funcMap).Parse(tmpl); err != nil {
		panic(err)
	}
	if err = tpl.Execute(&buf, struct {
		ConfigPackage string
		VoPackage     string
		DtoPackage    string
		SvcPackage    string
		Meta          astutils.InterfaceMeta
		Version       string
	}{
		DtoPackage:    dtoPkg,
		ConfigPackage: cfgPkg,
		SvcPackage:    ic.Package.Name,
		Meta:          meta,
		Version:       version.Release,
	}); err != nil {
		panic(err)
	}

	original, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	original = append(original, buf.Bytes()...)
	if tpl, err = template.New("simportimpl.go.tmpl").Parse(svcimportTmpl); err != nil {
		panic(err)
	}
	if err = tpl.Execute(&importBuf, struct {
		ConfigPackage string
		VoPackage     string
		DtoPackage    string
	}{
		DtoPackage:    dtoPkg,
		ConfigPackage: cfgPkg,
	}); err != nil {
		panic(err)
	}
	original = astutils.AppendImportStatements(original, importBuf.Bytes())
	original = astutils.RestRelatedModify(original, meta.Name)
	//fmt.Println(string(original))
	astutils.FixImport(original, svcimplfile)
}

func convert(m v3.Message) string {
	if !m.IsImported {
		return "pb." + m.String()
	}
	return m.String()
}

// GenSvcImplGrpc generates service implementation for grpc
func GenSvcImplGrpc(dir string, ic astutils.InterfaceCollector, grpcSvc v3.Service) {
	var (
		err         error
		svcimplfile string
		f           *os.File
		tpl         *template.Template
		buf         bytes.Buffer
		meta        astutils.InterfaceMeta
		tmpl        string
		importBuf   bytes.Buffer
	)
	svcimplfile = filepath.Join(dir, "svcimpl.go")
	err = copier.DeepCopy(ic.Interfaces[0], &meta)
	if err != nil {
		panic(err)
	}
	if _, err = os.Stat(svcimplfile); os.IsNotExist(err) {
		if f, err = os.Create(svcimplfile); err != nil {
			panic(err)
		}
		defer f.Close()
		tmpl = svcimplTmplGrpc
	} else {
		logrus.Warningln("New content will be append to file svcimpl.go")
		if f, err = os.OpenFile(svcimplfile, os.O_APPEND, os.ModePerm); err != nil {
			panic(err)
		}
		defer f.Close()
		tmpl = appendPartGrpc

		sc := astutils.NewStructCollector(astutils.ExprString)
		astutils.CollectStructsInFolder(dir, sc)
		if implementations, exists := sc.Methods[meta.Name+"Impl"]; exists {
			var notimplemented []v3.Rpc
			for _, item := range grpcSvc.Rpcs {
				for _, implemented := range implementations {
					if item.Name == implemented.Name {
						goto L
					}
				}
				notimplemented = append(notimplemented, item)

			L:
			}

			grpcSvc.Rpcs = notimplemented
		}
	}

	funcMap := make(map[string]interface{})
	funcMap["toCamel"] = strcase.ToCamel
	funcMap["convert"] = convert
	if tpl, err = template.New("svcimpl.go.tmpl").Funcs(funcMap).Parse(tmpl); err != nil {
		panic(err)
	}
	cfgPkg := astutils.GetPkgPath(filepath.Join(dir, "config"))
	dtoPkg := astutils.GetPkgPath(filepath.Join(dir, "dto"))
	pbPkg := astutils.GetPkgPath(filepath.Join(dir, "transport", "grpc"))
	if err = tpl.Execute(&buf, struct {
		ConfigPackage string
		VoPackage     string
		DtoPackage    string
		PbPackage     string
		SvcPackage    string
		Meta          astutils.InterfaceMeta
		GrpcSvc       v3.Service
		Version       string
	}{
		DtoPackage:    dtoPkg,
		ConfigPackage: cfgPkg,
		PbPackage:     pbPkg,
		SvcPackage:    ic.Package.Name,
		Meta:          meta,
		GrpcSvc:       grpcSvc,
		Version:       version.Release,
	}); err != nil {
		panic(err)
	}

	original, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	original = append(original, buf.Bytes()...)
	if tpl, err = template.New("simportimpl.go.tmpl").Parse(svcimportTmplGrpc); err != nil {
		panic(err)
	}
	if err = tpl.Execute(&importBuf, struct {
		ConfigPackage string
		VoPackage     string
		DtoPackage    string
		PbPackage     string
	}{
		DtoPackage:    dtoPkg,
		ConfigPackage: cfgPkg,
		PbPackage:     pbPkg,
	}); err != nil {
		panic(err)
	}
	original = astutils.AppendImportStatements(original, importBuf.Bytes())
	original = astutils.GrpcRelatedModify(original, meta.Name, grpcSvc.Name)
	//fmt.Println(string(original))
	astutils.FixImport(original, svcimplfile)
}

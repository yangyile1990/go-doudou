package template

const NotEditMark = `
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
`

const EditMark = `
// Code generated by gorm.io/gen. YOU CAN EDIT.
// Code generated by gorm.io/gen. YOU CAN EDIT.
// Code generated by gorm.io/gen. YOU CAN EDIT.
`

const NotEditMarkForGDDShort = `// Code generated by gorm.io/gen for go-doudou. DO NOT EDIT.`

const EditMarkForGDD = `
// Code generated by gorm.io/gen for go-doudou. YOU CAN EDIT.
// Code generated by gorm.io/gen for go-doudou. YOU CAN EDIT.
// Code generated by gorm.io/gen for go-doudou. YOU CAN EDIT.
`

const Header = NotEditMark + `
package {{.Package}}

import(	
	{{range .ImportPkgPaths}}{{.}}` + "\n" + `{{end}}
)
`

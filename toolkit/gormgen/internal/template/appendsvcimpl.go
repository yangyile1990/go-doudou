package template

const AppendSvcImpl = `
// Post{{.ModelStructName}} {{.StructComment}}
` + NotEditMarkForGDDShort + `
func (receiver *{{.InterfaceName}}Impl) Post{{.ModelStructName}}(ctx context.Context, body dto.{{.ModelStructName}}) (data {{.PriKeyType}}, err error) {
	m := model.{{.ModelStructName}}(body)
	u := receiver.q.{{.ModelStructName}}
	err = errors.WithStack(u.WithContext(ctx).Create(&m))
	data = m.ID
	return
}

// Post{{.ModelStructName}}s {{.StructComment}}
` + NotEditMarkForGDDShort + `
func (receiver *{{.InterfaceName}}Impl) Post{{.ModelStructName}}s(ctx context.Context, body []dto.{{.ModelStructName}}) (data []{{.PriKeyType}}, err error) {
	list := make([]*model.{{.ModelStructName}}, 0, len(body))
	for _, item := range body {
		m := model.{{.ModelStructName}}(item)
		list = append(list, &m)
	}
	u := receiver.q.{{.ModelStructName}}
	if err = errors.WithStack(u.WithContext(ctx).Create(list...)); err != nil {
		return
	}
	data = make([]{{.PriKeyType}}, 0, len(list))
	for _, item := range list {
		data = append(data, item.ID)
	}
	return
}

// Get{{.ModelStructName}}_Id {{.StructComment}}
` + NotEditMarkForGDDShort + `
func (receiver *{{.InterfaceName}}Impl) Get{{.ModelStructName}}_Id(ctx context.Context, id {{.PriKeyType}}) (data dto.{{.ModelStructName}}, err error) {
	u := receiver.q.{{.ModelStructName}}
	m, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return dto.{{.ModelStructName}}{}, errors.WithStack(err)
	}
	return dto.{{.ModelStructName}}(*m), nil
}

// Put{{.ModelStructName}} {{.StructComment}}
` + NotEditMarkForGDDShort + `
func (receiver *{{.InterfaceName}}Impl) Put{{.ModelStructName}}(ctx context.Context, body dto.{{.ModelStructName}}) (err error) {
	m := model.{{.ModelStructName}}(body)
	u := receiver.q.{{.ModelStructName}}
	_, err = u.WithContext(ctx).Where(u.ID.Eq(body.ID)).Updates(m)
	return errors.WithStack(err)
}

// Delete{{.ModelStructName}}_Id {{.StructComment}}
` + NotEditMarkForGDDShort + `
func (receiver *{{.InterfaceName}}Impl) Delete{{.ModelStructName}}_Id(ctx context.Context, id {{.PriKeyType}}) (err error) {
	u := receiver.q.{{.ModelStructName}}
	_, err = u.WithContext(ctx).Where(u.ID.Eq(id)).Delete()
	return errors.WithStack(err)
}

// Get{{.ModelStructName}}s {{.StructComment}}
` + NotEditMarkForGDDShort + `
func (receiver *{{.InterfaceName}}Impl) Get{{.ModelStructName}}s(ctx context.Context, parameter dto.Parameter) (data dto.Page, err error) {
	paginated := receiver.pg.With(database.Db.Model(&model.{{.ModelStructName}}{})).Request(paginate.Parameter(parameter)).Response(&[]model.{{.ModelStructName}}{})
	data = dto.Page(paginated)
	return
}

`

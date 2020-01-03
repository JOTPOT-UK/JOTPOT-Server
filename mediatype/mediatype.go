package mediatype

import "strings"

type Type struct {
	Type, Subtype string
	Params        []Param
}

func NewType(typ, subtyp string) Type {
	return Type{
		Type:    typ,
		Subtype: subtyp,
	}
}

type Param struct {
	Name, Value string
}

func (t *Type) GetParamCaseSensitive(name string) string {
	for _, p := range t.Params {
		if p.Name == name {
			return p.Value
		}
	}
	return ""
}

func (t *Type) GetParam(name string) string {
	return t.GetParamCaseSensitive(strings.ToLower(name))
}

func (t *Type) SetParam(name, value string) {
	for _, p := range t.Params {
		if p.Name == name {
			p.Value = value
			return
		}
	}
	t.Params = append(t.Params, Param{
		Name:  name,
		Value: value,
	})
}

func (t *Type) String() string {
	str := t.MimeType()
	for _, p := range t.Params {
		str += "; " + p.String()
	}
	return str
}

func (t *Type) MimeType() string {
	return t.Type + "/" + t.Subtype
}

func (p *Param) String() string {
	return p.Name + "=" + p.Value
}

package xsd

import (
	"encoding/xml"

	"github.com/iancoleman/strcase"
)

type Type interface {
	GoName() string
	GoTypeName() string
	Schema() *Schema
	Attributes() []Attribute
	Elements() []Element
	ContainsText() bool
	compile(*Schema, *Element)
}

type ComplexType struct {
	XMLName          xml.Name        `xml:"http://www.w3.org/2001/XMLSchema complexType"`
	Name             string          `xml:"name,attr"`
	Mixed            bool            `xml:"mixed,attr"`
	AttributesDirect []Attribute     `xml:"attribute"`
	Sequence         *Sequence       `xml:"sequence"`
	schema           *Schema         `xml:"-"`
	SimpleContent    *SimpleContent  `xml:"simpleContent"`
	ComplexContent   *ComplexContent `xml:"complexContent"`
	Choice           *Choice         `xml:"choice"`
	content          GenericContent  `xml:"-"`
}

func (ct *ComplexType) Attributes() []Attribute {
	if ct.content != nil {
		return ct.content.Attributes()
	}
	return ct.AttributesDirect
}

func (ct *ComplexType) Elements() []Element {
	if ct.Sequence != nil {
		return ct.Sequence.Elements()
	} else if ct.content != nil {
		return ct.content.Elements()
	} else if ct.Choice != nil {
		return ct.Choice.Elements()
	}
	return []Element{}
}

func (ct *ComplexType) GoName() string {
	return strcase.ToCamel(ct.Name)
}

func (ct *ComplexType) GoTypeName() string {
	return ct.GoName()
}

func (ct *ComplexType) ContainsText() bool {
	return ct.content != nil && ct.content.ContainsText()
}

func (ct *ComplexType) Schema() *Schema {
	return ct.schema
}

func (ct *ComplexType) compile(sch *Schema, parentElement *Element) {
	ct.schema = sch
	if ct.Sequence != nil {
		ct.Sequence.compile(sch, parentElement)
	}

	// Handle improbable name clash. Consider XSD defining two attributes on the element:
	// "id" and "Id", this would create name clash given the camelization we do.
	goNames := map[string]uint{}
	for idx, _ := range ct.Attributes() {
		attribute := &ct.Attributes()[idx]
		attribute.compile(sch)

		count := goNames[attribute.GoName()]
		count += 1
		goNames[attribute.GoName()] = count
		attribute.DuplicateCount = count
		// Second GoName may be different depending on the DuplicateCount
		goNames[attribute.GoName()] = count
	}

	if ct.ComplexContent != nil {
		ct.content = ct.ComplexContent
		if ct.SimpleContent != nil {
			panic("Not implemented: xsd:complexType " + ct.Name + " defines xsd:simpleContent and xsd:complexContent together")
		}
	} else if ct.SimpleContent != nil {
		ct.content = ct.SimpleContent
	}

	if ct.content != nil {
		if len(ct.AttributesDirect) > 1 {
			panic("Not implemented: xsd:complexType " + ct.Name + " defines direct attribute and xsd:*Content")
		}
		if ct.Sequence != nil {
			panic("Not implemented: xsd:complexType " + ct.Name + " defines xsd:sequence and xsd:*Content")
		}
		ct.content.compile(sch, parentElement)
	}

	if ct.Choice != nil {
		if ct.content != nil {
			panic("Not implemented: xsd:complexType " + ct.Name + " defines xsd:choice and xsd:*content")
		}
		if ct.Sequence != nil {
			panic("Not implemented: xsd:complexType " + ct.Name + " defines xsd:choice and xsd:sequence")
		}
		ct.Choice.compile(sch, parentElement)
	}
}

type SimpleType struct {
	XMLName xml.Name `xml:"http://www.w3.org/2001/XMLSchema simpleType"`
	Name    string   `xml:"name,attr"`
	schema  *Schema  `xml:"-"`
}

func (st *SimpleType) GoName() string {
	return strcase.ToCamel(st.Name)
}

func (st *SimpleType) GoTypeName() string {
	return "string"
}

func (st *SimpleType) Schema() *Schema {
	return st.schema
}

func (st *SimpleType) compile(sch *Schema, parentElement *Element) {
	st.schema = sch
}

func (st *SimpleType) Attributes() []Attribute {
	return []Attribute{}
}

func (st *SimpleType) Elements() []Element {
	return []Element{}
}

func (st *SimpleType) ContainsText() bool {
	return true
}

type staticType string

func (st staticType) GoName() string {
	return string(st)
}

func (ct staticType) GoTypeName() string {
	return ct.GoName()
}

func (st staticType) Attributes() []Attribute {
	return []Attribute{}
}

func (st staticType) Elements() []Element {
	return []Element{}
}

func (st staticType) Schema() *Schema {
	return nil
}

func (staticType) ContainsText() bool {
	return true
}

func (st staticType) compile(*Schema, *Element) {
}

var staticTypes = map[string]staticType{
	"string":           "string",
	"dateTime":         "string",
	"base64Binary":     "string",
	"normalizedString": "string",
	"token":            "string",
	"NCName":           "string",
	"anySimpleType":    "string",
	"int":              "int",
	"integer":          "int64",
	"decimal":          "float64",
	"boolean":          "bool",
}

func StaticType(name string) staticType {
	typ, found := staticTypes[name]
	if found {
		return typ
	}
	panic("Type xsd:" + name + " not implemented")
	return staticType(name)
}

func IsStaticType(name string) bool {
	_, found := staticTypes[name]
	return found
}

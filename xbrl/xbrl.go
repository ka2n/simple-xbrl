package xbrl

import (
	"encoding/xml"
	"io"
	"time"
)

// XBRL instance
type XBRL struct {
	XMLName  xml.Name
	Contexts []Context `xml:"context"`
	Facts    []Fact    `xml:",any"`
}

// Context represents <ixbrl:context> tag
type Context struct {
	XMLName xml.Name
	ID      string `xml:"id,attr"`
	Instant Date   `xml:"period>instant"`
	Start   Date   `xml:"period>startDate"`
	End     Date   `xml:"period>endDate"`
}

// Fact represents each fact
type Fact struct {
	XMLName    xml.Name
	Name       string
	Value      string
	ContextRef string
	UnitRef    string
	Decimals   string
	Nil        bool
}

// UnmarshalXML implements xml.Unmarshaler interface
func (t *Fact) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v struct {
		XMLName    xml.Name
		Name       string `xml:"name,attr"`
		Value      string `xml:",chardata"`
		ContextRef string `xml:"contextRef,attr"`
		UnitRef    string `xml:"unitRef,attr"`
		Decimals   string `xml:"decimals,attr"`
		Nil        string `xml:"nil,attr"`
	}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}

	// parse xsi:nil
	isnil := v.Nil == "true"

	// use name attribute if exists, or use XML tag's local name
	var name string
	if v.Name != "" {
		name = v.Name
	} else {
		name = v.XMLName.Local
	}

	*t = Fact{v.XMLName, name, v.Value, v.ContextRef, v.UnitRef, v.Decimals, isnil}
	return nil
}

// Date container
type Date struct {
	time.Time
}

// UnmarshalXML implements xml.Unmarshaler interface
func (t *Date) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	err := d.DecodeElement(&v, &start)
	if err != nil {
		return err
	}

	const form = "2006-01-02"
	tt, err := time.Parse(form, v)
	if err != nil {
		return err
	}

	*t = Date{tt}
	return nil
}

// UnmarshalXBRL unmarshal io.Reader into *XBRL or return error
func UnmarshalXBRL(xbrl *XBRL, reader io.Reader) error {
	decoder := xml.NewDecoder(reader)
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			// Ingore HTML tags
			if se.Name.Space == "http://www.w3.org/1999/xhtml" {
				continue
			}
			switch n := se.Name.Local; n {
			case "xbrl": // Vanila XBRL
				if err := decoder.DecodeElement(&xbrl, &se); err != nil {
					return err
				}
				return nil
			// inlineXBRL
			case "hidden", "resources", "references", "unit", "header":
				continue
			case "context":
				var ctx Context
				if err := decoder.DecodeElement(&ctx, &se); err != nil {
					return err
				}
				xbrl.Contexts = append(xbrl.Contexts, ctx)
			default:
				var fact Fact
				if err := decoder.DecodeElement(&fact, &se); err != nil {
					return err
				}
				xbrl.Facts = append(xbrl.Facts, fact)
			}
		}
	}
	return nil
}

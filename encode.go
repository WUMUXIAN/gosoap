package gosoap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"reflect"
)

var tokens []xml.Token

// MarshalXML envelope the body and encode to xml
func (c Client) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {

	tokens = []xml.Token{}

	//start envelope
	if c.Definitions == nil {
		return fmt.Errorf("definitions is nil")
	}

	startEnvelope()
	if len(c.HeaderParams) > 0 {
		startHeader(c.HeaderName, c.Definitions.Types[0].XsdSchema[0].TargetNamespace)
		for k, v := range c.HeaderParams {
			t := xml.StartElement{
				Name: xml.Name{
					Space: "",
					Local: k,
				},
			}

			tokens = append(tokens, t, xml.CharData(v), xml.EndElement{Name: t.Name})
		}

		endHeader(c.HeaderName)
	}

	err := startBody(c.Method, c.Definitions.Types[0].XsdSchema[0].TargetNamespace)
	if err != nil {
		return err
	}

	if len(c.Params) != 0 {
		recursiveEncode(c.Params)
	} else {
		encodeBody(c.StructParam)
	}

	//end envelope
	endBody(c.Method)
	endEnvelope()

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	return e.Flush()
}

func encodeBody(s interface{}) {
	content, _ := xml.Marshal(s)
	decoder := xml.NewDecoder(bytes.NewReader(content))
	// Let's skip the first start element.
	decoder.Token()
	nameSpace := ""
	nameSpaceTokenName := ""
	for {
		if token, e := decoder.Token(); e == nil {
			switch v := token.(type) {
			case xml.StartElement:
				startElement := xml.StartElement{
					Name: xml.Name{
						Space: "",
						Local: v.Name.Local,
					},
					Attr: v.Attr,
				}
				// Starting of a namespace
				if nameSpace == "" && v.Name.Space != "" {
					nameSpace = v.Name.Space
					nameSpaceTokenName = v.Name.Local
					startElement.Name.Space = nameSpace
				}
				tokens = append(tokens, startElement)
			case xml.EndElement:
				endElemennt := xml.EndElement{
					Name: xml.Name{
						Space: "",
						Local: v.Name.Local,
					},
				}
				// Closing of a namespace
				if nameSpaceTokenName == v.Name.Local {
					endElemennt.Name.Space = nameSpace
					nameSpace = ""
					nameSpaceTokenName = ""
				}
				tokens = append(tokens, endElemennt)
			case xml.CharData:
				tokens = append(tokens, xml.CharData(string(v)))
			default:
				tokens = append(tokens, token)
			}
		} else {
			break
		}
	}
	// Let's remove the last element.
	tokens = tokens[:len(tokens)-1]
}

func recursiveEncode(hm interface{}) {
	v := reflect.ValueOf(hm)

	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			t := xml.StartElement{
				Name: xml.Name{
					Space: "",
					Local: key.String(),
				},
			}

			tokens = append(tokens, t)
			recursiveEncode(v.MapIndex(key).Interface())
			tokens = append(tokens, xml.EndElement{Name: t.Name})
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			recursiveEncode(v.Index(i).Interface())
		}
	case reflect.String:
		content := xml.CharData(v.String())
		tokens = append(tokens, content)
	}
}

func startEnvelope() {
	e := xml.StartElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Envelope",
		},
		Attr: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "xmlns:xsi"}, Value: "http://www.w3.org/2001/XMLSchema-instance"},
			{Name: xml.Name{Space: "", Local: "xmlns:xsd"}, Value: "http://www.w3.org/2001/XMLSchema"},
			{Name: xml.Name{Space: "", Local: "xmlns:soap"}, Value: "http://schemas.xmlsoap.org/soap/envelope/"},
		},
	}

	tokens = append(tokens, e)
}

func endEnvelope() {
	e := xml.EndElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Envelope",
		},
	}

	tokens = append(tokens, e)
}

func startHeader(m, n string) {
	h := xml.StartElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Header",
		},
	}

	if m == "" || n == "" {
		tokens = append(tokens, h)
		return
	}

	r := xml.StartElement{
		Name: xml.Name{
			Space: "",
			Local: m,
		},
		Attr: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "xmlns"}, Value: n},
		},
	}

	tokens = append(tokens, h, r)

	return
}

func endHeader(m string) {
	h := xml.EndElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Header",
		},
	}

	if m == "" {
		tokens = append(tokens, h)
		return
	}

	r := xml.EndElement{
		Name: xml.Name{
			Space: "",
			Local: m,
		},
	}

	tokens = append(tokens, r, h)
}

// startToken initiate body of the envelope
func startBody(m, n string) error {
	b := xml.StartElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Body",
		},
	}

	if m == "" || n == "" {
		return fmt.Errorf("method or namespace is empty")
	}

	r := xml.StartElement{
		Name: xml.Name{
			Space: "",
			Local: m,
		},
		Attr: []xml.Attr{
			{Name: xml.Name{Space: "", Local: "xmlns"}, Value: n},
		},
	}

	tokens = append(tokens, b, r)

	return nil
}

// endToken close body of the envelope
func endBody(m string) {
	b := xml.EndElement{
		Name: xml.Name{
			Space: "",
			Local: "soap:Body",
		},
	}

	r := xml.EndElement{
		Name: xml.Name{
			Space: "",
			Local: m,
		},
	}

	tokens = append(tokens, r, b)
}

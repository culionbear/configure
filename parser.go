package configure

import (
	"encoding/json"
	"encoding/xml"
	"gopkg.in/yaml.v3"
)

type Parser interface {
	Marshal(value any) ([]byte, error)
	Unmarshal(config any, buf []byte) error
}

var (
	Json Parser = new(JsonParser)
	Yaml Parser = new(YamlParser)
	Xml  Parser = new(XmlParser)
)

type JsonParser struct{}

func (p *JsonParser) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (p *JsonParser) Unmarshal(config any, buf []byte) error {
	return json.Unmarshal(buf, config)
}

type YamlParser struct{}

func (p *YamlParser) Marshal(value any) ([]byte, error) {
	return yaml.Marshal(value)
}

func (p *YamlParser) Unmarshal(config any, buf []byte) error {
	return yaml.Unmarshal(buf, config)
}

type XmlParser struct{}

func (p *XmlParser) Marshal(value any) ([]byte, error) {
	return xml.Marshal(value)
}

func (p *XmlParser) Unmarshal(config any, buf []byte) error {
	return xml.Unmarshal(buf, config)
}

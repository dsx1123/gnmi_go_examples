package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/gnxi/utils/xpath"
	"github.com/openconfig/gnmi/proto/gnmi"
)

type Output struct {
	Path  string                 `json:"path"`
	Value map[string]interface{} `json:"value"`
}

func PrettyJSON(d *[]byte) *bytes.Buffer {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, *d, "", "    ")
	if err != nil {
		_ = fmt.Errorf("Parse json error: %s", err)
	}
	return &prettyJSON

}

func GetXPath(p *gnmi.Path) string {
	var xpath string
	for _, pElem := range p.GetElem() {
		xpath = xpath + "/" + pElem.Name
		if pElem.Key != nil {
			xpath = xpath + parseMapKey(pElem.GetKey())
		}
	}
	return xpath
}

func parseMapKey(m map[string]string) string {
	var keys []string
	for k := range m {
		keys = append(keys, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return fmt.Sprintf("[%s]", strings.Join(keys, ","))
}

func ParsePath(p string) (string, *gnmi.Path, error) {
	var origin string

	if len(p) == 0 {
		return origin, &gnmi.Path{}, nil
	}

	oIndex := strings.Index(p, ":") // find end index of origin
	if oIndex >= 0 {
		origin = p[:oIndex]
		p = p[oIndex+1:]
	}
	pElem, err := xpath.ToGNMIPath(p)
	if err != nil {
		fmt.Println(err)
		return origin, nil, err
	}

	return origin, pElem, nil
}

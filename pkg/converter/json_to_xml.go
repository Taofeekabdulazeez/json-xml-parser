package converter

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// xmlNode represents an intermediate node in our XML structure
type xmlNode struct {
	Name       string
	Attributes []xml.Attr
	Children   []*xmlNode
	Content    string
}

// marshal writes the xmlNode and its children to the XML encoder
func (n *xmlNode) marshal(enc *xml.Encoder) error {
	start := xml.StartElement{
		Name: xml.Name{Local: n.Name},
		Attr: n.Attributes,
	}

	if err := enc.EncodeToken(start); err != nil {
		return err
	}

	if n.Content != "" {
		if err := enc.EncodeToken(xml.CharData(n.Content)); err != nil {
			return err
		}
	}

	for _, child := range n.Children {
		if err := child.marshal(enc); err != nil {
			return err
		}
	}

	return enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: n.Name}})
}

// JSONToXML converts JSON bytes into XML bytes using the provided options.
func JSONToXML(jsonBytes []byte, opts JSONToXMLOptions) ([]byte, error) {
	// Set default options if not provided
	if opts.RootElement == "" {
		opts.RootElement = "root"
	}
	if opts.AttributePrefix == "" {
		opts.AttributePrefix = "@"
	}
	if opts.ArrayMode == "" {
		opts.ArrayMode = ArrayModeWrapper
	}

	// Use json.NewDecoder with UseNumber to preserve exact number formatting
	var data interface{}
	decoder := json.NewDecoder(bytes.NewReader(jsonBytes))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("malformed JSON: %w", err)
	}

	rootName := SanitizeXMLName(opts.RootElement)

	switch v := data.(type) {
	case map[string]interface{}:
		rootNode := &xmlNode{Name: rootName}

		// Sort keys to guarantee deterministic XML structure (useful for testing & readability)
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			val := v[k]
			if strings.HasPrefix(k, opts.AttributePrefix) {
				attrName := strings.TrimPrefix(k, opts.AttributePrefix)
				rootNode.Attributes = append(rootNode.Attributes, xml.Attr{
					Name:  xml.Name{Local: SanitizeXMLName(attrName)},
					Value: formatPrimitive(val),
				})
			} else if k == "#text" || k == "$" {
				rootNode.Content = formatPrimitive(val)
			} else {
				children := buildXMLTree(k, val, opts)
				rootNode.Children = append(rootNode.Children, children...)
			}
		}
		return serializeXML(rootNode, opts.PrettyPrint)

	case []interface{}:
		rootNode := &xmlNode{Name: rootName}
		for _, item := range v {
			children := buildXMLTree("item", item, opts)
			rootNode.Children = append(rootNode.Children, children...)
		}
		return serializeXML(rootNode, opts.PrettyPrint)

	default:
		rootNode := &xmlNode{
			Name:    rootName,
			Content: formatPrimitive(v),
		}
		return serializeXML(rootNode, opts.PrettyPrint)
	}
}

// buildXMLTree recursively constructs the XML node tree from parsed JSON values
func buildXMLTree(name string, val interface{}, opts JSONToXMLOptions) []*xmlNode {
	sanitizedName := SanitizeXMLName(name)

	switch v := val.(type) {
	case map[string]interface{}:
		node := &xmlNode{Name: sanitizedName}

		// Sort keys to maintain deterministic tag ordering
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			childVal := v[k]
			if strings.HasPrefix(k, opts.AttributePrefix) {
				attrName := strings.TrimPrefix(k, opts.AttributePrefix)
				node.Attributes = append(node.Attributes, xml.Attr{
					Name:  xml.Name{Local: SanitizeXMLName(attrName)},
					Value: formatPrimitive(childVal),
				})
			} else if k == "#text" || k == "$" {
				node.Content = formatPrimitive(childVal)
			} else {
				children := buildXMLTree(k, childVal, opts)
				node.Children = append(node.Children, children...)
			}
		}
		return []*xmlNode{node}

	case []interface{}:
		if opts.ArrayMode == ArrayModeWrapper {
			node := &xmlNode{Name: sanitizedName}
			for _, item := range v {
				children := buildXMLTree("item", item, opts)
				node.Children = append(node.Children, children...)
			}
			return []*xmlNode{node}
		} else {
			var nodes []*xmlNode
			for _, item := range v {
				children := buildXMLTree(name, item, opts)
				nodes = append(nodes, children...)
			}
			return nodes
		}

	default:
		node := &xmlNode{
			Name:    sanitizedName,
			Content: formatPrimitive(val),
		}
		return []*xmlNode{node}
	}
}

// formatPrimitive returns the string representation of basic types
func formatPrimitive(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	case json.Number:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

// serializeXML serializes the XML tree to a byte slice
func serializeXML(root *xmlNode, prettyPrint bool) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	enc := xml.NewEncoder(&buf)
	if prettyPrint {
		enc.Indent("", "  ")
	}

	if err := root.marshal(enc); err != nil {
		return nil, fmt.Errorf("failed to encode XML tokens: %w", err)
	}

	if err := enc.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush XML encoder: %w", err)
	}

	if prettyPrint {
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}

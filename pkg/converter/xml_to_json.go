package converter

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// XMLNode represents a node in the parsed XML tree.
type XMLNode struct {
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Children   []*XMLNode        `json:"children,omitempty"`
	Content    string            `json:"content,omitempty"`
}

// XMLToJSON converts XML bytes into JSON bytes using the provided options.
func XMLToJSON(xmlBytes []byte, opts XMLToJSONOptions) ([]byte, error) {
	// Set default options if not provided
	if opts.AttributePrefix == "" {
		opts.AttributePrefix = "@"
	}
	if opts.TextKey == "" {
		opts.TextKey = "#text"
	}

	// Parse XML to a tree
	rootNode, err := parseXMLTree(xmlBytes)
	if err != nil {
		return nil, fmt.Errorf("malformed XML: %w", err)
	}

	// Clean tree formatting whitespace
	cleanTree(rootNode)

	// Convert tree to JSON-compatible structure
	converted := nodeToJSON(rootNode, opts)

	// Wrap in root key to preserve root element name (symmetrical parsing)
	resultMap := map[string]interface{}{
		rootNode.Name: converted,
	}

	var output []byte
	if opts.PrettyPrint {
		output, err = json.MarshalIndent(resultMap, "", "  ")
	} else {
		output, err = json.Marshal(resultMap)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return output, nil
}

// parseXMLTree reads XML tokens and builds a hierarchical tree of XMLNodes.
func parseXMLTree(xmlBytes []byte) (*XMLNode, error) {
	dec := xml.NewDecoder(bytes.NewReader(xmlBytes))
	var stack []*XMLNode
	var root *XMLNode

	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch se := t.(type) {
		case xml.StartElement:
			node := &XMLNode{
				Name:       se.Name.Local,
				Attributes: make(map[string]string),
			}
			for _, attr := range se.Attr {
				node.Attributes[attr.Name.Local] = attr.Value
			}

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			} else {
				if root == nil {
					root = node
				}
			}
			stack = append(stack, node)

		case xml.CharData:
			if len(stack) > 0 {
				curr := stack[len(stack)-1]
				curr.Content += string(se)
			}

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}

	if root == nil {
		return nil, errors.New("empty or invalid XML document")
	}

	return root, nil
}

// cleanTree removes formatting whitespace around XML elements.
func cleanTree(node *XMLNode) {
	if len(node.Children) > 0 {
		// If there are children, any direct text content is formatting whitespace
		node.Content = ""
	} else {
		// If it's a leaf node, we keep the content but trim whitespace if it's purely empty space.
		trimmed := strings.TrimSpace(node.Content)
		if trimmed == "" {
			node.Content = ""
		}
	}

	for _, child := range node.Children {
		cleanTree(child)
	}
}

// nodeToJSON recursively converts an XMLNode to a JSON-compatible type.
func nodeToJSON(node *XMLNode, opts XMLToJSONOptions) interface{} {
	// Case 1: Simple leaf node with no attributes
	if len(node.Attributes) == 0 && len(node.Children) == 0 {
		if opts.AutoTypeConversion {
			return parseAutoType(node.Content)
		}
		return node.Content
	}

	// Case 2: Node has attributes and/or children (becomes a JSON object)
	result := make(map[string]interface{})

	// Add attributes
	for k, v := range node.Attributes {
		keyName := opts.AttributePrefix + k
		if opts.AutoTypeConversion {
			result[keyName] = parseAutoType(v)
		} else {
			result[keyName] = v
		}
	}

	// Add content text if any exists
	if node.Content != "" {
		if opts.AutoTypeConversion {
			result[opts.TextKey] = parseAutoType(node.Content)
		} else {
			result[opts.TextKey] = node.Content
		}
	}

	// Group children by name to detect and handle arrays
	childGroups := make(map[string][]*XMLNode)
	var childNames []string // Keep sorted list of names for determinism

	for _, child := range node.Children {
		if _, exists := childGroups[child.Name]; !exists {
			childNames = append(childNames, child.Name)
		}
		childGroups[child.Name] = append(childGroups[child.Name], child)
	}
	sort.Strings(childNames)

	for _, name := range childNames {
		group := childGroups[name]
		if len(group) == 1 {
			// Single child
			result[name] = nodeToJSON(group[0], opts)
		} else {
			// Multiple children with same tag name = Array
			var list []interface{}
			for _, itemNode := range group {
				list = append(list, nodeToJSON(itemNode, opts))
			}
			result[name] = list
		}
	}

	return result
}

// parseAutoType attempts to convert string values to bool, int, or float
func parseAutoType(val string) interface{} {
	if val == "" {
		return ""
	}

	// Boolean checks
	if val == "true" {
		return true
	}
	if val == "false" {
		return false
	}

	// Integer with leading zero check (e.g. zip codes "02138" or phone codes "01" should remain strings)
	if len(val) > 1 && val[0] == '0' && val[1] != '.' && (val[1] >= '0' && val[1] <= '9') {
		return val
	}
	// Negative integer with leading zero check (e.g. "-0123")
	if len(val) > 2 && val[0] == '-' && val[1] == '0' && val[2] != '.' && (val[2] >= '0' && val[2] <= '9') {
		return val
	}

	// Try integer parsing
	if i, err := strconv.ParseInt(val, 10, 64); err == nil {
		return i
	}

	// Try float parsing
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	}

	return val
}

package converter

import (
	"regexp"
	"strings"
)

// ArrayMode determines how slices/arrays are represented in the generated XML.
type ArrayMode string

const (
	// ArrayModeWrapper wraps array items in a parent tag and uses a generic or customized name for items.
	// Example: {"skills": ["Go", "JS"]} -> <skills><item>Go</item><item>JS</item></skills>
	ArrayModeWrapper ArrayMode = "wrapper"

	// ArrayModeRepeating repeats the element name for each item in the array without a wrapper.
	// Example: {"skills": ["Go", "JS"]} -> <skills>Go</skills><skills>JS</skills>
	ArrayModeRepeating ArrayMode = "repeating"
)

// JSONToXMLOptions holds the options for converting JSON to XML.
type JSONToXMLOptions struct {
	RootElement     string    `json:"root_element"`     // Name of the root XML element. Defaults to "root".
	ArrayMode       ArrayMode `json:"array_mode"`       // How arrays are handled. Defaults to "wrapper".
	AttributePrefix string    `json:"attribute_prefix"` // Prefix for JSON keys that represent XML attributes. Defaults to "@".
	PrettyPrint     bool      `json:"pretty_print"`     // If true, output will be indented. Defaults to true.
}

// XMLToJSONOptions holds the options for converting XML to JSON.
type XMLToJSONOptions struct {
	AttributePrefix    string `json:"attribute_prefix"`    // Prefix for XML attributes when mapped to JSON keys. Defaults to "@".
	TextKey            string `json:"text_key"`            // JSON key name for text content in mixed nodes. Defaults to "#text".
	AutoTypeConversion bool   `json:"auto_type_conversion"` // If true, attempts to parse numeric and boolean values from strings. Defaults to true.
	PrettyPrint        bool   `json:"pretty_print"`        // If true, JSON output is pretty-printed. Defaults to true.
}

var (
	invalidXMLStartChar = regexp.MustCompile(`^[^a-zA-Z_]`)
	invalidXMLChars      = regexp.MustCompile(`[^a-zA-Z0-9_\-\.]`)
)

// SanitizeXMLName converts an arbitrary string into a valid XML tag/attribute name.
// XML naming rules:
// - Must start with a letter or underscore
// - Cannot start with "xml" (or XML, Xml, etc.)
// - Can contain letters, digits, hyphens, underscores, and periods
// - Cannot contain spaces or other symbols
func SanitizeXMLName(name string) string {
	if name == "" {
		return "item"
	}

	// Remove leading/trailing spaces
	name = strings.TrimSpace(name)

	// Replace spaces and special characters with underscores
	sanitized := invalidXMLChars.ReplaceAllString(name, "_")

	// If it starts with an invalid character, prepend an underscore
	if invalidXMLStartChar.MatchString(sanitized) {
		sanitized = "_" + sanitized
	}

	// Check if it starts with 'xml' (reserved prefix)
	if len(sanitized) >= 3 && strings.ToLower(sanitized[:3]) == "xml" {
		sanitized = "_" + sanitized
	}

	return sanitized
}

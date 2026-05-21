package converter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeXMLName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "item"},
		{"name", "name"},
		{"first name", "first_name"},
		{"123name", "_123name"},
		{"xmlName", "_xmlName"},
		{"XML-Name", "_XML-Name"},
		{"valid-name.123", "valid-name.123"},
		{"invalid#char@here", "invalid_char_here"},
	}

	for _, tc := range tests {
		got := SanitizeXMLName(tc.input)
		if got != tc.expected {
			t.Errorf("SanitizeXMLName(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestJSONToXML(t *testing.T) {
	jsonInput := `{
		"@id": 42,
		"name": "Jane Doe",
		"active": true,
		"skills": ["Go", "Docker"],
		"address": {
			"city": "Boston",
			"zip": "02108"
		}
	}`

	t.Run("ArrayModeWrapper", func(t *testing.T) {
		opts := JSONToXMLOptions{
			RootElement:     "user",
			ArrayMode:       ArrayModeWrapper,
			AttributePrefix: "@",
			PrettyPrint:     false,
		}
		xmlBytes, err := JSONToXML([]byte(jsonInput), opts)
		if err != nil {
			t.Fatalf("JSONToXML error: %v", err)
		}

		xmlStr := string(xmlBytes)
		expectedParts := []string{
			`<user id="42">`,
			`<name>Jane Doe</name>`,
			`<active>true</active>`,
			`<skills><item>Go</item><item>Docker</item></skills>`,
			`<address><city>Boston</city><zip>02108</zip></address>`,
			`</user>`,
		}

		for _, part := range expectedParts {
			if !strings.Contains(xmlStr, part) {
				t.Errorf("XML missing segment: %s\nGot:\n%s", part, xmlStr)
			}
		}
	})

	t.Run("ArrayModeRepeating", func(t *testing.T) {
		opts := JSONToXMLOptions{
			RootElement:     "user",
			ArrayMode:       ArrayModeRepeating,
			AttributePrefix: "@",
			PrettyPrint:     false,
		}
		xmlBytes, err := JSONToXML([]byte(jsonInput), opts)
		if err != nil {
			t.Fatalf("JSONToXML error: %v", err)
		}

		xmlStr := string(xmlBytes)
		// Skills elements should be consecutive repeating tags, not wrapped in a <skills> block
		if !strings.Contains(xmlStr, "<skills>Go</skills><skills>Docker</skills>") {
			t.Errorf("XML missing repeating arrays:\nGot:\n%s", xmlStr)
		}
	})
}

func TestXMLToJSON(t *testing.T) {
	xmlInput := `<?xml version="1.0" encoding="UTF-8"?>
	<user id="42">
		<name>Jane Doe</name>
		<active>true</active>
		<skills>Go</skills>
		<skills>Docker</skills>
		<address>
			<city>Boston</city>
			<zip>02108</zip>
		</address>
	</user>`

	t.Run("WithAutoTypeAndPrefix", func(t *testing.T) {
		opts := XMLToJSONOptions{
			AttributePrefix:    "@",
			TextKey:            "#text",
			AutoTypeConversion: true,
			PrettyPrint:        false,
		}

		jsonBytes, err := XMLToJSON([]byte(xmlInput), opts)
		if err != nil {
			t.Fatalf("XMLToJSON error: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			t.Fatalf("unmarshal result JSON error: %v", err)
		}

		userMap, ok := result["user"].(map[string]interface{})
		if !ok {
			t.Fatalf("JSON root is not 'user' object: %v", result)
		}

		// Verify type conversions
		// Note that unmarshaling JSON numbers back to interfaces converts them to float64 by default
		if idVal, exists := userMap["@id"]; !exists {
			t.Errorf("expected @id attribute to exist")
		} else if f, ok := idVal.(float64); !ok || f != 42 {
			t.Errorf("expected @id to be 42, got %T: %v", idVal, idVal)
		}

		if userMap["active"] != true {
			t.Errorf("expected active to be true, got %v", userMap["active"])
		}

		// Verify array grouping
		skills, ok := userMap["skills"].([]interface{})
		if !ok || len(skills) != 2 || skills[0] != "Go" || skills[1] != "Docker" {
			t.Errorf("expected skills array ['Go', 'Docker'], got %v", userMap["skills"])
		}

		// Verify zip leading zero preservation
		address, ok := userMap["address"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected address map, got %v", userMap["address"])
		}
		if address["zip"] != "02108" {
			t.Errorf("expected zip to remain string '02108', got %T: %v", address["zip"], address["zip"])
		}
	})
}

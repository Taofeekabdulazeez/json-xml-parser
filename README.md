# Transforma - JSON & XML Bidirectional Parser

`Transforma` is a senior-grade, high-performance web service and interactive API written in Go using the **Gin** framework. It provides robust bidirectional conversion between JSON and XML documents, complete with customizable parsing options and a premium, responsive glassmorphism UI.

## Features

- **Bidirectional Live Conversion**: Interactive split-pane UI that converts JSON to XML (left-to-right) or XML to JSON (right-to-left) in real-time as you type or paste.
- **XML Tag Name Sanitization**: Auto-resolves invalid XML character keys (spaces, numbers, special characters) to maintain schema compliance.
- **Customizable Array Formats**:
  - `wrapper` mode: Groups array elements inside a parent tag (e.g. `<skills><item>Go</item></skills>`).
  - `repeating` mode: Repeats elements directly without a parent tag (e.g. `<skills>Go</skills><skills>Docker</skills>`).
- **Flexible Attributes Mapping**: Supports mapping JSON keys starting with a specific prefix (default `@`) to XML attributes on elements.
- **Intelligent Auto-Type Detection**: Optionally parses numeric strings and boolean tags in XML to actual JSON numbers and booleans, while preserving formatting like leading zeros for zip codes or codes (e.g. `"02108"` remains a string).
- **RESTful API**: Standard REST endpoints for programmatic access.
- **Premium UI Aesthetics**: Responsive, glowing glassmorphic interface built using HSL colors, modern typography, line-number sync, and modal drawer API docs.

---

## Project Structure

```
json-xml-parser/
├── main.go               # Entry point of the Gin web server & routes
├── go.mod                # Go module definition
├── docs/
│   └── openapi.yaml      # OpenAPI 3.0 API Specification
├── pkg/
│   └── converter/        # Core conversion package
│       ├── converter.go  # Core types, configurations, and name sanitizers
│       ├── json_to_xml.go# JSON to XML recursive builder
│       ├── xml_to_json.go# XML to JSON token parser
│       └── converter_test.go # Comprehensive unit tests
├── static/
│   ├── css/
│   │   └── style.css     # Premium UI theme stylesheets (glassmorphism)
│   └── js/
│       └── app.js        # Event handling, scroll sync, and debounced fetch requests
└── templates/
    └── index.html        # Go HTML UI layout template
```

---

## Installation & Setup

### Prerequisites

- Go `1.22` or later (tested on `go1.26.2`).

### Running the Application

1. Clone or navigate to the workspace directory.
2. Initialize and tidy modules:
   ```bash
   go mod tidy
   ```
3. Run the development server:
   ```bash
   go run main.go
   ```
4. Access the UI at: **[http://localhost:8080](http://localhost:8080)**

### Running Unit Tests

To run the test suite for the conversion package:
```bash
go test -v ./pkg/converter/...
```

---

## API Documentation

The server exposes the following REST endpoints:

### 1. JSON to XML
- **Endpoint**: `POST /api/v1/convert/json-to-xml`
- **Headers**: `Content-Type: application/json`
- **Query Options**:
  - `root`: (string, default: `root`) XML root element name.
  - `array_mode`: (string, `wrapper` or `repeating`, default: `wrapper`) Format for lists.
  - `attr_prefix`: (string, default: `@`) Indicator prefix for XML attributes.
  - `pretty`: (boolean, default: `true`) Pretty-prints formatting.

### 2. XML to JSON
- **Endpoint**: `POST /api/v1/convert/xml-to-json`
- **Headers**: `Content-Type: application/xml`
- **Query Options**:
  - `attr_prefix`: (string, default: `@`) Indicator prefix for XML attributes.
  - `text_key`: (string, default: `#text`) Key for element text when mixed with attributes.
  - `auto_type`: (boolean, default: `true`) Converts numbers/booleans from strings.
  - `pretty`: (boolean, default: `true`) Indents JSON output.

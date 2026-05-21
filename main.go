package main

import (
	"io"
	"net/http"

	"json-xml-parser/pkg/converter"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to release mode if needed, or leave default for debugging
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Serve static files
	r.Static("/static", "./static")

	// HTML UI route
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "JSON & XML Bidirectional Parser",
		})
	})

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		v1.POST("/convert/json-to-xml", convertJSONToXML)
		v1.POST("/convert/xml-to-json", convertXMLToJSON)
	}

	// Start server on port 8080
	r.Run(":8080")
}

// convertJSONToXML handles POST requests containing a raw JSON body,
// parses conversion settings from URL query parameters, converts it,
// and returns the formatted XML document.
func convertJSONToXML(c *gin.Context) {
	// Read raw body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body: " + err.Error(),
		})
		return
	}

	if len(bodyBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request body is empty",
		})
		return
	}

	// Parse query parameters
	opts := converter.JSONToXMLOptions{
		RootElement:     c.DefaultQuery("root", "root"),
		ArrayMode:       converter.ArrayMode(c.DefaultQuery("array_mode", "wrapper")),
		AttributePrefix: c.DefaultQuery("attr_prefix", "@"),
		PrettyPrint:     c.DefaultQuery("pretty", "true") != "false",
	}

	// Validate array mode
	if opts.ArrayMode != converter.ArrayModeWrapper && opts.ArrayMode != converter.ArrayModeRepeating {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid array_mode: must be 'wrapper' or 'repeating'",
		})
		return
	}

	// Perform conversion
	xmlBytes, err := converter.JSONToXML(bodyBytes, opts)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Respond with XML header and body
	c.Data(http.StatusOK, "application/xml; charset=utf-8", xmlBytes)
}

// convertXMLToJSON handles POST requests containing a raw XML body,
// parses conversion settings from URL query parameters, converts it,
// and returns the formatted JSON document.
func convertXMLToJSON(c *gin.Context) {
	// Read raw body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body: " + err.Error(),
		})
		return
	}

	if len(bodyBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request body is empty",
		})
		return
	}

	// Parse query parameters
	opts := converter.XMLToJSONOptions{
		AttributePrefix:    c.DefaultQuery("attr_prefix", "@"),
		TextKey:            c.DefaultQuery("text_key", "#text"),
		AutoTypeConversion: c.DefaultQuery("auto_type", "true") != "false",
		PrettyPrint:        c.DefaultQuery("pretty", "true") != "false",
	}

	// Perform conversion
	jsonBytes, err := converter.XMLToJSON(bodyBytes, opts)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Respond with JSON
	c.Data(http.StatusOK, "application/json; charset=utf-8", jsonBytes)
}

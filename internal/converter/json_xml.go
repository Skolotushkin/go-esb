package converter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

// JSONToXML конвертирует JSON в XML
func JSONToXML(jsonData []byte) ([]byte, error) {
	var jsonObj interface{}
	if err := json.Unmarshal(jsonData, &jsonObj); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	xmlData := convertToXML(jsonObj, "root")
	output, err := xml.MarshalIndent(xmlData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML declaration
	return []byte(xml.Header + string(output)), nil
}

// XMLToJSON конвертирует XML в JSON
func XMLToJSON(xmlData []byte) ([]byte, error) {
	var xmlObj XMLNode
	if err := xml.Unmarshal(xmlData, &xmlObj); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	jsonObj := convertToJSON(xmlObj)
	output, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return output, nil
}

// XMLNode представляет XML элемент
type XMLNode struct {
	XMLName xml.Name
	Content string     `xml:",chardata"`
	Attrs   []xml.Attr `xml:",attr"`
	Child   []XMLNode  `xml:",any"`
}

func convertToXML(obj interface{}, rootName string) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		elements := []xml.Name{}
		for key, _ := range v {
			xmlElem := xml.Name{Local: sanitizeXMLName(key)}
			elements = append(elements, xmlElem)
		}
		// Простая структура для XML
		return map[string]interface{}{
			"root": v,
		}
	case []interface{}:
		return map[string]interface{}{
			"items": v,
		}
	default:
		return map[string]interface{}{
			rootName: v,
		}
	}
}

func convertToJSON(xmlNode XMLNode) interface{} {
	if len(xmlNode.Child) == 0 {
		// Leaf node
		if xmlNode.Content != "" {
			return xmlNode.Content
		}
		if len(xmlNode.Attrs) > 0 {
			result := make(map[string]interface{})
			for _, attr := range xmlNode.Attrs {
				result["@"+attr.Name.Local] = attr.Value
			}
			return result
		}
		return nil
	}

	// Node with children
	result := make(map[string]interface{})
	
	// Process attributes
	for _, attr := range xmlNode.Attrs {
		result["@"+attr.Name.Local] = attr.Value
	}

	// Process children
	childMap := make(map[string]interface{})
	for _, child := range xmlNode.Child {
		key := child.XMLName.Local
		value := convertToJSON(child)
		
		if existing, exists := childMap[key]; exists {
			// Convert to array if duplicate keys
			if arr, ok := existing.([]interface{}); ok {
				childMap[key] = append(arr, value)
			} else {
				childMap[key] = []interface{}{existing, value}
			}
		} else {
			childMap[key] = value
		}
	}

	if xmlNode.Content != "" && strings.TrimSpace(xmlNode.Content) != "" {
		result["#text"] = xmlNode.Content
	}

	if len(childMap) > 0 {
		for k, v := range childMap {
			result[k] = v
		}
	}

	return result
}

func sanitizeXMLName(name string) string {
	// Replace invalid XML characters
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	if len(name) > 0 && (name[0] >= '0' && name[0] <= '9') {
		name = "n" + name
	}
	return name
}


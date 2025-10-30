package converter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
)

// FormatConverter интерфейс для конвертации между форматами
type FormatConverter interface {
	Convert(data []byte, fromFormat, toFormat string) ([]byte, error)
}

// Converter реализует FormatConverter
type Converter struct{}

// NewConverter создает новый конвертер
func NewConverter() *Converter {
	return &Converter{}
}

// Convert конвертирует данные между форматами
func (c *Converter) Convert(data []byte, fromFormat, toFormat string) ([]byte, error) {
	if fromFormat == toFormat {
		return data, nil
	}

	// JSON ↔ XML
	if (fromFormat == "JSON" && toFormat == "XML") || (fromFormat == "XML" && toFormat == "JSON") {
		if fromFormat == "JSON" {
			return JSONToXML(data)
		}
		return XMLToJSON(data)
	}

	// JSON ↔ CSV
	if fromFormat == "JSON" && toFormat == "CSV" {
		return JSONToCSV(data)
	}
	if fromFormat == "CSV" && toFormat == "JSON" {
		return CSVToJSON(data)
	}

	// XML ↔ CSV (через JSON)
	if fromFormat == "XML" && toFormat == "CSV" {
		jsonData, err := XMLToJSON(data)
		if err != nil {
			return nil, err
		}
		return JSONToCSV(jsonData)
	}
	if fromFormat == "CSV" && toFormat == "XML" {
		jsonData, err := CSVToJSON(data)
		if err != nil {
			return nil, err
		}
		return JSONToXML(jsonData)
	}

	return nil, fmt.Errorf("unsupported conversion: %s -> %s", fromFormat, toFormat)
}

// JSONToCSV конвертирует JSON в CSV
func JSONToCSV(jsonData []byte) ([]byte, error) {
	var jsonObj interface{}
	if err := json.Unmarshal(jsonData, &jsonObj); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var rows [][]string
	var headers []string
	headerSet := make(map[string]bool)

	// Flatten JSON object to rows
	flattenJSON(jsonObj, &rows, &headers, headerSet, []string{})

	if len(rows) == 0 {
		return []byte{}, nil
	}

	// Ensure all rows have same length
	maxCols := len(headers)
	for i := range rows {
		for len(rows[i]) < maxCols {
			rows[i] = append(rows[i], "")
		}
	}

	var buf strings.Builder
	writer := csv.NewWriter(&buf)
	
	// Write headers
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Write rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(buf.String()), nil
}

// CSVToJSON конвертирует CSV в JSON
func CSVToJSON(csvData []byte) ([]byte, error) {
	reader := csv.NewReader(strings.NewReader(string(csvData)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return []byte("[]"), nil
	}

	headers := records[0]
	var result []map[string]string

	for i := 1; i < len(records); i++ {
		row := make(map[string]string)
		for j, header := range headers {
			if j < len(records[i]) {
				row[header] = records[i][j]
			} else {
				row[header] = ""
			}
		}
		result = append(result, row)
	}

	return json.Marshal(result)
}

func flattenJSON(obj interface{}, rows *[][]string, headers *[]string, headerSet map[string]bool, path []string) {
	switch v := obj.(type) {
	case map[string]interface{}:
		row := make([]string, len(*headers))
		rowMap := make(map[string]string)

		for key, value := range v {
			currentPath := append(path, key)
			pathKey := strings.Join(currentPath, ".")
			
			switch val := value.(type) {
			case map[string]interface{}:
				flattenJSON(val, rows, headers, headerSet, currentPath)
			case []interface{}:
				for i, item := range val {
					arrayPath := append(currentPath, fmt.Sprintf("%d", i))
					flattenJSON(item, rows, headers, headerSet, arrayPath)
				}
			default:
				if !headerSet[pathKey] {
					*headers = append(*headers, pathKey)
					headerSet[pathKey] = true
				}
				rowMap[pathKey] = fmt.Sprintf("%v", val)
			}
		}

		if len(rowMap) > 0 {
			*rows = append(*rows, make([]string, len(*headers)))
			for i, h := range *headers {
				(*rows)[len(*rows)-1][i] = rowMap[h]
			}
		}

	case []interface{}:
		for _, item := range v {
			flattenJSON(item, rows, headers, headerSet, path)
		}

	default:
		pathKey := strings.Join(path, ".")
		if !headerSet[pathKey] {
			*headers = append(*headers, pathKey)
			headerSet[pathKey] = true
		}
		row := make([]string, len(*headers))
		for i, h := range *headers {
			if h == pathKey {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		*rows = append(*rows, row)
	}
}


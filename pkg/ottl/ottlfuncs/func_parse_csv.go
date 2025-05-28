// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottlfuncs // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottlfuncs"

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/parseutils"
	// "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/rwxdex/opentelemetry-collector-contrib-custom/pkg/ottl"
)

const (
	parseCSVModeStrict       = "strict"
	parseCSVModeLazyQuotes   = "lazyQuotes"
	parseCSVModeIgnoreQuotes = "ignoreQuotes"
)

const (
	parseCSVDefaultDelimiter = ','
	parseCSVDefaultMode      = parseCSVModeStrict
)

type ParseCSVArguments[K any] struct {
	Target          ottl.StringGetter[K]
	Header          ottl.Optional[string]   // Опционально, если заголовки берутся из файла
	Delimiter       ottl.Optional[string]
	HeaderDelimiter ottl.Optional[string]
	// Mode            ottl.Optional[string]
	FilePath        ottl.Optional[string]   // Путь к файлу data.csv
}

func (p ParseCSVArguments[K]) validate() error {
	if !p.Delimiter.IsEmpty() {
		if len([]rune(p.Delimiter.Get())) != 1 {
			return errors.New("delimiter must be a single character")
		}
	}

	if !p.HeaderDelimiter.IsEmpty() {
		if len([]rune(p.HeaderDelimiter.Get())) != 1 {
			return errors.New("header_delimiter must be a single character")
		}
	}

	return nil
}

func NewParseCSVFactory[K any]() ottl.Factory[K] {
	return ottl.NewFactory("ParseCSV", &ParseCSVArguments[K]{}, createParseCSVFunction[K])
}

func createParseCSVFunction[K any](_ ottl.FunctionContext, oArgs ottl.Arguments) (ottl.ExprFunc[K], error) {
	args, ok := oArgs.(*ParseCSVArguments[K])
	if !ok {
		return nil, errors.New("ParseCSVFactory args must be of type *ParseCSVArguments[K]")
	}

	if err := args.validate(); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	delimiter := parseCSVDefaultDelimiter
	if !args.Delimiter.IsEmpty() {
		delimiter = []rune(args.Delimiter.Get())[0]
	}

	// headerDelimiter defaults to the chosen delimiter,
	// since in most cases headerDelimiter == delimiter.
	headerDelimiter := string(delimiter)
	if !args.HeaderDelimiter.IsEmpty() {
		headerDelimiter = args.HeaderDelimiter.Get()
	}

	// mode := parseCSVDefaultMode
	// if !args.Mode.IsEmpty() {
	// 	mode = args.Mode.Get()
	// }

	filePath := "/otelcol/data.csv" // Значение по умолчанию
	if !args.FilePath.IsEmpty() {
		filePath = args.FilePath.Get()
	}

	var parseRow parseCSVRowFunc
	switch mode {
	case parseCSVModeStrict:
		parseRow = parseCSVRow(false)
	case parseCSVModeLazyQuotes:
		parseRow = parseCSVRow(true)
	case parseCSVModeIgnoreQuotes:
		parseRow = parseCSVRowIgnoreQuotes()
	default:
		return nil, fmt.Errorf("unknown mode: %s", mode)
	}

	// Предварительно загружаем данные из файла
	csvData, headers, err := loadCSVFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error loading CSV file: %w", err)
	}

	// Если заголовки заданы явно, используем их вместо заголовков из файла
	if !args.Header.IsEmpty() {
		headers = strings.Split(args.Header.Get(), headerDelimiter)
	}

	return parseCSVWithData[K](args.Target, headers, delimiter, parseRow, csvData), nil
}

// Структура для хранения предварительно загруженных данных CSV
type csvFileData struct {
	records [][]string
	headers []string
}

// Загружает CSV файл и возвращает его содержимое
func loadCSVFile(filePath string) ([][]string, []string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Читаем заголовки
	headers, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading headers: %w", err)
	}

	// Читаем все записи
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading CSV data: %w", err)
	}

	return records, headers, nil
}

type parseCSVRowFunc func(row string, delimiter rune) ([]string, error)

// Модифицированная функция parseCSV, которая использует предварительно загруженные данные
func parseCSVWithData[K any](target ottl.StringGetter[K], headers []string, delimiter rune, parseRow parseCSVRowFunc, csvData [][]string) ottl.ExprFunc[K] {
	return func(ctx context.Context, tCtx K) (any, error) {
		targetStr, err := target.Get(ctx, tCtx)
		if err != nil {
			return nil, fmt.Errorf("error getting value for target in ParseCSV: %w", err)
		}

		// Поиск строки по ключу в первом столбце
		var fields []string
		keyFound := false
		
		for _, record := range csvData {
			if len(record) > 0 && record[0] == targetStr {
				fields = record
				keyFound = true
				break
			}
		}

		// Если строка не найдена, парсим targetStr как CSV строку
		if !keyFound {
			fields, err = parseRow(targetStr, delimiter)
			if err != nil {
				return nil, err
			}
		}

		headersToFields, err := mapCSVHeaders(headers, fields)
		if err != nil {
			return nil, fmt.Errorf("map csv headers: %w", err)
		}

		pMap := pcommon.NewMap()
		err = pMap.FromRaw(headersToFields)
		return pMap, err
	}
}

// Сопоставляет заголовки и значения
func mapCSVHeaders(headers, fields []string) (map[string]interface{}, error) {
	if len(fields) > len(headers) {
		return nil, fmt.Errorf("number of fields (%d) exceeds number of headers (%d)", len(fields), len(headers))
	}

	headersToFields := make(map[string]interface{})
	
	for i, header := range headers {
		if i < len(fields) {
			// Удаляем пробелы с начала и конца
			header = strings.TrimSpace(header)
			if header != "" {
				headersToFields[header] = fields[i]
			}
		}
	}

	return headersToFields, nil
}

func parseCSVRow(lazyQuotes bool) parseCSVRowFunc {
	return func(row string, delimiter rune) ([]string, error) {
		r := csv.NewReader(strings.NewReader(row))
		r.Comma = delimiter
		r.LazyQuotes = lazyQuotes
		
		fields, err := r.Read()
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error parsing CSV row: %w", err)
		}
		
		return fields, nil
	}
}

func parseCSVRowIgnoreQuotes() parseCSVRowFunc {
	return func(row string, delimiter rune) ([]string, error) {
		return strings.Split(row, string([]rune{delimiter})), nil
	}
}

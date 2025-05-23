// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottlfuncs // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottlfuncs"

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

const (
	parseCSVModeStrict       = "strict"
	parseCSVModeLazyQuotes   = "lazyQuotes"
	parseCSVModeIgnoreQuotes = "ignoreQuotes"
)

const (
	parseCSVDefaultDelimiter = ','
	parseCSVDefaultMode      = parseCSVModeStrict
	parseCSVFileReloadInterval = 5 * time.Minute // Интервал перезагрузки файла
)

// Структура для хранения данных CSV файла
type csvFileCache struct {
	filePath     string
	headers      []string
	records      [][]string
	recordsMap   map[string][]string // Для быстрого поиска по ключу
	lastModified time.Time
	mutex        sync.RWMutex
	logger       *zap.Logger
}

// Глобальный кэш файлов CSV для повторного использования
var csvFileCaches = make(map[string]*csvFileCache)
var csvFileCachesMutex sync.Mutex

// Получение или создание кэша для файла
func getOrCreateCSVFileCache(filePath string, logger *zap.Logger) (*csvFileCache, error) {
	csvFileCachesMutex.Lock()
	defer csvFileCachesMutex.Unlock()

	if cache, ok := csvFileCaches[filePath]; ok {
		return cache, nil
	}

	cache := &csvFileCache{
		filePath:   filePath,
		recordsMap: make(map[string][]string),
		logger:     logger,
	}

	// Загружаем данные из файла
	if err := cache.reload(); err != nil {
		return nil, err
	}

	// Запускаем горутину для периодической перезагрузки файла
	go cache.periodicReload()

	csvFileCaches[filePath] = cache
	return cache, nil
}

// Перезагрузка данных из файла
func (c *csvFileCache) reload() error {
	fileInfo, err := os.Stat(c.filePath)
	if err != nil {
		return fmt.Errorf("error checking file %s: %w", c.filePath, err)
	}

	// Проверяем, изменился ли файл
	if !fileInfo.ModTime().After(c.lastModified) && len(c.records) > 0 {
		return nil // Файл не изменился
	}

	file, err := os.Open(c.filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", c.filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Читаем заголовки
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("error reading headers from %s: %w", c.filePath, err)
	}

	// Читаем все записи
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading records from %s: %w", c.filePath, err)
	}

	// Создаем карту для быстрого поиска
	recordsMap := make(map[string][]string, len(records))
	for _, record := range records {
		if len(record) > 0 {
			recordsMap[record[0]] = record
		}
	}

	// Обновляем данные атомарно
	c.mutex.Lock()
	c.headers = headers
	c.records = records
	c.recordsMap = recordsMap
	c.lastModified = fileInfo.ModTime()
	c.mutex.Unlock()

	c.logger.Info("CSV file loaded", 
		zap.String("file", c.filePath), 
		zap.Int("records", len(records)),
		zap.Time("lastModified", fileInfo.ModTime()))

	return nil
}

// Периодическая перезагрузка файла
func (c *csvFileCache) periodicReload() {
	ticker := time.NewTicker(parseCSVFileReloadInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.reload(); err != nil {
			c.logger.Warn("Failed to reload CSV file", 
				zap.String("file", c.filePath), 
				zap.Error(err))
		}
	}
}

// Поиск записи по ключу
func (c *csvFileCache) findRecord(key string) ([]string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	record, found := c.recordsMap[key]
	return record, found
}

// Получение заголовков
func (c *csvFileCache) getHeaders() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Создаем копию заголовков
	headers := make([]string, len(c.headers))
	copy(headers, c.headers)
	return headers
}

// Аргументы для функции ParseCSVFile
type ParseCSVFileArguments[K any] struct {
	Target    ottl.StringGetter[K]
	FilePath  ottl.StringGetter[K]
	Delimiter ottl.Optional[string]
	Mode      ottl.Optional[string]
}

func (p ParseCSVFileArguments[K]) validate() error {
	if !p.Delimiter.IsEmpty() {
		if len([]rune(p.Delimiter.Get())) != 1 {
			return errors.New("delimiter must be a single character")
		}
	}
	return nil
}

// Фабрика для создания функции ParseCSVFile
func NewParseCSVFileFactory[K any](logger *zap.Logger) ottl.Factory[K] {
	return ottl.NewFactory("ParseCSVFile", &ParseCSVFileArguments[K]{}, func(ctx ottl.FunctionContext, oArgs ottl.Arguments) (ottl.ExprFunc[K], error) {
		return createParseCSVFileFunction[K](ctx, oArgs, logger)
	})
}

// Создание функции ParseCSVFile
func createParseCSVFileFunction[K any](_ ottl.FunctionContext, oArgs ottl.Arguments, logger *zap.Logger) (ottl.ExprFunc[K], error) {
	args, ok := oArgs.(*ParseCSVFileArguments[K])
	if !ok {
		return nil, errors.New("ParseCSVFileFactory args must be of type *ParseCSVFileArguments[K]")
	}

	if err := args.validate(); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	delimiter := parseCSVDefaultDelimiter
	if !args.Delimiter.IsEmpty() {
		delimiter = []rune(args.Delimiter.Get())[0]
	}

	mode := parseCSVDefaultMode
	if !args.Mode.IsEmpty() {
		mode = args.Mode.Get()
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

	return func(ctx context.Context, tCtx K) (any, error) {
		targetStr, err := args.Target.Get(ctx, tCtx)
		if err != nil {
			return nil, fmt.Errorf("error getting value for target in ParseCSVFile: %w", err)
		}

		filePathStr, err := args.FilePath.Get(ctx, tCtx)
		if err != nil {
			return nil, fmt.Errorf("error getting value for filePath in ParseCSVFile: %w", err)
		}

		// Получаем кэш для файла
		cache, err := getOrCreateCSVFileCache(filePathStr, logger)
		if err != nil {
			return nil, fmt.Errorf("error loading CSV file: %w", err)
		}

		// Ищем запись по ключу
		record, found := cache.findRecord(targetStr)
		if !found {
			// Если запись не найдена, пробуем разобрать targetStr как CSV строку
			record, err = parseRow(targetStr, delimiter)
			if err != nil {
				return nil, fmt.Errorf("key '%s' not found in CSV file and failed to parse as CSV: %w", targetStr, err)
			}
		}

		// Получаем заголовки
		headers := cache.getHeaders()

		// Создаем карту заголовок -> значение
		headersToFields, err := mapCSVHeaders(headers, record)
		if err != nil {
			return nil, fmt.Errorf("map csv headers: %w", err)
		}

		pMap := pcommon.NewMap()
		err = pMap.FromRaw(headersToFields)
		return pMap, err
	}, nil
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

type parseCSVRowFunc func(row string, delimiter rune) ([]string, error)

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

package httpapi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"MonitorPeople/internal/domain"

	xlsreader "github.com/shakinm/xlsReader/xls"
	"github.com/xuri/excelize/v2"
)

var (
	errUnsupportedImportFile = errors.New("unsupported import file type")
	errImportColumnsNotFound = errors.New("required import columns not found")
)

func parseStudentsImport(reader io.Reader, filename string) ([]domain.PersonDraft, error) {
	rows, err := readWorkbookRows(reader, filename)
	if err != nil {
		return nil, err
	}

	for _, sheetRows := range rows {
		headerRow, fioCol, englishFIOCol, programCol := findStudentColumns(sheetRows)
		if fioCol < 0 || programCol < 0 {
			continue
		}

		drafts := make([]domain.PersonDraft, 0, len(sheetRows)-headerRow-1)
		for _, row := range sheetRows[headerRow+1:] {
			fio := studentFIOFromRow(row, fioCol, englishFIOCol)
			program := sanitizeCell(cellAt(row, programCol))
			surname, name, ok := splitFullName(fio)
			if !ok || strings.TrimSpace(program) == "" {
				continue
			}
			drafts = append(drafts, domain.PersonDraft{
				Name:           name,
				Surname:        surname,
				StudyDirection: cleanStudyProgram(program),
			})
		}
		return drafts, nil
	}

	return nil, errImportColumnsNotFound
}

func parseTeachersImport(reader io.Reader, filename string) ([]domain.PersonDraft, error) {
	rows, err := readWorkbookRows(reader, filename)
	if err != nil {
		return nil, err
	}

	for _, sheetRows := range rows {
		headerRow, fioCol := findTeacherFIOColumn(sheetRows)
		if fioCol < 0 {
			continue
		}

		drafts := make([]domain.PersonDraft, 0, len(sheetRows)-headerRow-1)
		for _, row := range sheetRows[headerRow+1:] {
			surname, name, ok := splitFullName(sanitizeCell(cellAt(row, fioCol)))
			if !ok {
				continue
			}
			drafts = append(drafts, domain.PersonDraft{
				Name:           name,
				Surname:        surname,
				StudyDirection: domain.TeacherStudyDirection,
			})
		}
		return drafts, nil
	}

	return nil, errImportColumnsNotFound
}

func readWorkbookRows(reader io.Reader, filename string) ([][][]string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read import file: %w", err)
	}

	switch ext {
	case ".xlsx":
		return readXLSXRows(bytes.NewReader(data))
	case ".xls":
		return readXLSRows(bytes.NewReader(data))
	default:
		return nil, errUnsupportedImportFile
	}
}

func readXLSXRows(reader io.Reader) ([][][]string, error) {
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	allRows := make([][][]string, 0)
	for _, sheetName := range file.GetSheetList() {
		rows, err := file.GetRows(sheetName)
		if err != nil {
			return nil, err
		}
		allRows = append(allRows, trimEmptyRows(rows))
	}
	return allRows, nil
}

func readXLSRows(reader io.ReadSeeker) ([][][]string, error) {
	workbook, err := xlsreader.OpenReader(reader)
	if err != nil {
		return nil, err
	}

	allRows := make([][][]string, 0, workbook.GetNumberSheets())
	for sheetIndex := 0; sheetIndex < workbook.GetNumberSheets(); sheetIndex++ {
		sheet, err := workbook.GetSheet(sheetIndex)
		if err != nil {
			return nil, err
		}

		rows := make([][]string, 0, len(sheet.GetRows()))
		for _, row := range sheet.GetRows() {
			values := make([]string, 0)
			for _, col := range row.GetCols() {
				values = append(values, sanitizeCell(col.GetString()))
			}
			rows = append(rows, values)
		}
		allRows = append(allRows, trimEmptyRows(rows))
	}

	return allRows, nil
}

func findStudentColumns(rows [][]string) (int, int, int, int) {
	for rowIndex, row := range rows {
		fioCol := -1
		englishFIOCol := -1
		programCol := -1
		for colIndex, value := range row {
			normalized := normalizeHeader(value)
			if normalized == "фио" {
				fioCol = colIndex
			}
			if normalized == "фионаанглийском" {
				englishFIOCol = colIndex
			}
			if normalized == "образовательнаяпрограмма" {
				programCol = colIndex
			}
		}
		if fioCol >= 0 && programCol >= 0 {
			return rowIndex, fioCol, englishFIOCol, programCol
		}
	}
	return -1, -1, -1, -1
}

func findTeacherFIOColumn(rows [][]string) (int, int) {
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			normalized := normalizeHeader(value)
			if normalized == "укажитевашефио" || normalized == "фио" {
				return rowIndex, colIndex
			}
		}
	}
	return -1, -1
}

func splitFullName(fullName string) (string, string, bool) {
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) < 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func studentFIOFromRow(row []string, fioCol, englishFIOCol int) string {
	fio := sanitizeCell(cellAt(row, fioCol))
	englishFIO := sanitizeCell(cellAt(row, englishFIOCol))

	if hasCyrillic(fio) {
		if hasCyrillic(englishFIO) {
			return fio + englishFIO
		}
		return fio
	}

	if hasCyrillic(englishFIO) {
		return englishFIO
	}

	return fio
}

func hasCyrillic(value string) bool {
	for _, char := range value {
		if char >= 'А' && char <= 'я' || char == 'Ё' || char == 'ё' {
			return true
		}
	}
	return false
}

func cleanStudyProgram(program string) string {
	program = sanitizeCell(program)
	for _, marker := range []string{"очно-заочная ", "очная ", "заочная "} {
		if index := strings.Index(program, marker); index >= 0 {
			cleaned := strings.TrimSpace(program[index+len(marker):])
			if cleaned != "" {
				return cleaned
			}
		}
	}
	return program
}

func sanitizeCell(value string) string {
	value = strings.Map(func(char rune) rune {
		if char == 0 {
			return -1
		}
		if char < 32 && char != '\t' && char != '\n' && char != '\r' {
			return -1
		}
		return char
	}, value)
	return strings.TrimSpace(value)
}

func normalizeHeader(value string) string {
	value = strings.ToLower(sanitizeCell(value))
	value = strings.ReplaceAll(value, "ё", "е")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func cellAt(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return sanitizeCell(row[index])
}

func trimEmptyRows(rows [][]string) [][]string {
	for len(rows) > 0 && isEmptyRow(rows[len(rows)-1]) {
		rows = rows[:len(rows)-1]
	}
	return rows
}

func isEmptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

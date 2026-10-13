package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	model "core/internal/model"
)

type RowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type ParseResult struct {
	Companies []model.Company
	Errors    []RowError
}

func Parse(reader io.Reader) (ParseResult, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1

	header, err := csvReader.Read()
	if err != nil {
		return ParseResult{}, fmt.Errorf("read csv header: %w", err)
	}

	index := mapHeaders(header)
	if index.name < 0 {
		return ParseResult{}, fmt.Errorf(
			"csv must include a company name column (name, company or company_name)",
		)
	}

	result := ParseResult{}
	rowNum := 1

	for {
		rowNum++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, RowError{
				Row:     rowNum,
				Message: err.Error(),
			})
			continue
		}

		name := fieldAt(record, index.name)
		if name == "" {
			result.Errors = append(result.Errors, RowError{
				Row:     rowNum,
				Message: "missing company name",
			})
			continue
		}

		company := model.Company{
			Name:            name,
			Domain:          NormalizeDomain(fieldAt(record, index.domain)),
			Industry:        fieldAt(record, index.industry),
			Location:        fieldAt(record, index.location),
			Description:     fieldAt(record, index.description),
			RevenueEstimate: fieldAt(record, index.revenue),
			OwnerName:       fieldAt(record, index.owner),
			Employees:       intAt(record, index.employees),
			FoundedYear:     intAt(record, index.founded),
			Source:          model.SourceCSV,
		}
		result.Companies = append(result.Companies, company)
	}

	return result, nil
}

type headerIndex struct {
	name        int
	domain      int
	industry    int
	location    int
	description int
	employees   int
	revenue     int
	founded     int
	owner       int
}

func mapHeaders(header []string) headerIndex {
	idx := headerIndex{-1, -1, -1, -1, -1, -1, -1, -1, -1}
	for i, col := range header {
		switch normalizeHeader(col) {
		case "name", "company", "companyname", "business", "businessname":
			idx.name = i
		case "domain", "website", "url", "site":
			idx.domain = i
		case "industry", "sector", "vertical":
			idx.industry = i
		case "location", "country", "city", "region", "geography", "hq":
			idx.location = i
		case "description", "summary", "about":
			idx.description = i
		case "employees", "employeecount", "headcount", "staff":
			idx.employees = i
		case "revenue", "revenueestimate", "estimatedrevenue", "arr", "turnover":
			idx.revenue = i
		case "founded", "foundedyear", "yearfounded":
			idx.founded = i
		case "owner", "ownername", "founder", "contact":
			idx.owner = i
		}
	}
	return idx
}

func normalizeHeader(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "\ufeff")
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}

func fieldAt(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func intAt(record []string, index int) *int {
	value := strings.ReplaceAll(fieldAt(record, index), ",", "")
	if value == "" {
		return nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return nil
	}
	return &parsed
}

func NormalizeDomain(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Hostname() == "" {
		return ""
	}
	return strings.TrimPrefix(parsed.Hostname(), "www.")
}

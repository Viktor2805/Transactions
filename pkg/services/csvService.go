package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
)

type CSVService struct{}

type CSVServiceI interface {
	ReadCSVHeaders(reader *csv.Reader) ([]string, error)
	ReadCSVRecord(reader *csv.Reader) ([]string, error)
	GetCSVHeaders(model interface{}) []string
}

func NewCSVService() *CSVService {
	return &CSVService{}
}

func (s *CSVService) ReadCSVHeaders(reader *csv.Reader) ([]string, error) {
	record, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return []string{}, nil
		}
		return nil, fmt.Errorf("error reading CSV headers: %v", err)
	}
	return record, nil
}

func (s *CSVService) ReadCSVRecord(reader *csv.Reader) ([]string, error) {
	record, err := reader.Read()
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file: %v", err)
	}
	return record, nil
}

func (s *CSVService) GetCSVHeaders(model interface{}) []string {
	val := reflect.TypeOf(model)
	headers := make([]string, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		headers[i] = val.Field(i).Tag.Get("csv")
	}

	return headers
}

package protoparser

import (
	"os"
	"testing"
)

func TestParseSimpleFile(t *testing.T) {
	reader, err := os.Open("_testdata/simple.proto")
	if err != nil {
		t.Error("Failed to open file")
	}
	defer reader.Close()
	_, err = Parse(reader)
	if err != nil {
		t.Errorf("Failed to parse proto, %v", err)
	}
}

func TestParseTMEnumFile(t *testing.T) {
	reader, err := os.Open("_testdata/tmEnum.proto")
	if err != nil {
		t.Error("Failed to open file")
	}
	defer reader.Close()
	_, err = Parse(reader)
	if err != nil {
		t.Errorf("Failed to parse proto, %v", err)
	}
}

func TestParseReservedEnumFile(t *testing.T) {
	reader, err := os.Open("_testdata/reservedEnum.proto")
	if err != nil {
		t.Error("Failed to open file")
	}
	defer reader.Close()
	_, err = Parse(reader)
	if err != nil {
		t.Errorf("Failed to parse proto, %v", err)
	}
}

func TestParseReleaseVersionFile(t *testing.T) {
	reader, err := os.Open("_testdata/releaseVersion.proto")
	if err != nil {
		t.Error("Failed to open file")
	}
	defer reader.Close()
	_, err = Parse(reader)
	if err != nil {
		t.Errorf("Failed to parse proto, %v", err)
	}
}

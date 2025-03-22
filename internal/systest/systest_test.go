package systest

import (
	"embed"
	"testing"
)

//go:embed testdata/*
var testdata embed.FS

func Test_noreqs(t *testing.T) {
	RunSysTest(t, testdata, "noreqs", []string{"trace"}, "0.0.1")
}

func Test_err_undetected(t *testing.T) {
	RunSysTest(t, testdata, "err_undetected", []string{"trace"}, "0.0.1")
}

package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"go.uber.org/nilaway"
	"go.uber.org/nilaway/config"
	"gopkg.in/yaml.v3"
	"honnef.co/go/tools/staticcheck"

	"github.com/postman17/metrics/cmd/staticlint/analyzers"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
)

const Config = `config.yaml`

type ConfigData struct {
	CheckPrefix string   `yaml:"check_prefix"`
	Staticcheck []string `yaml:"static_check"`
}

func readConfig() ([]byte, error) {
	appfile, err := os.Executable()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err == nil {
		return data, nil
	}
	return os.ReadFile(Config)
}

func main() {
	data, err := readConfig()
	if err != nil {
		panic(err)
	}
	var cfg ConfigData
	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		panic(err)
	}
	mychecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		errcheck.Analyzer,
		nilaway.Analyzer,
		analyzers.NoOsExitAnalyzer,
	}
	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, cfg.CheckPrefix) || checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	if err := config.Analyzer.Flags.Set(config.ExcludePkgsFlag, "compress/gzip,go/parser,go/printer,go/types,go/ast,go/doc,go.uber.org/nilaway,golang.org/x/tools"); err != nil {
		panic(err)
	}
	multichecker.Main(
		mychecks...,
	)
}

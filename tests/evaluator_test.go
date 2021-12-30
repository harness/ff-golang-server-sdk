package tests

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/harness/ff-golang-server-sdk/log"
	"github.com/harness/ff-golang-server-sdk/pkg/evaluation"
	"github.com/harness/ff-golang-server-sdk/pkg/repository"
	"github.com/harness/ff-golang-server-sdk/rest"
)

const source = "./ff-test-cases/tests"

type testFile struct {
	Filename string
	Flag     rest.FeatureConfig     `json:"flag"`
	Segments []rest.Segment         `json:"segments"`
	Targets  []rest.Target          `json:"targets"`
	Expected map[string]interface{} `json:"expected"`
}

func loadFiles() []testFile {
	files, err := ioutil.ReadDir(source)
	if err != nil {
		log.Error(err)
	}

	slice := make([]testFile, 0, len(files))
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		if f, err := loadFile(file.Name()); err == nil {
			slice = append(slice, f)
		}
	}
	return slice
}

func loadFile(filename string) (testFile, error) {
	fp := filepath.Clean(filepath.Join(source, filename))
	content, err := ioutil.ReadFile(fp)
	if err != nil {
		log.Error(err)
		return testFile{}, err
	}

	result := testFile{
		Filename: filename,
	}
	err = json.Unmarshal(content, &result)
	if err != nil {
		log.Error(err)
		return testFile{}, err
	}
	return result, nil
}

func TestEvaluator(t *testing.T) {
	type args struct {
		file     string
		target   string
		value    interface{}
		testFile testFile
	}

	tests := make([]args, 0)
	data := loadFiles()
	lruCache, err := repository.NewLruCache(1000)
	if err != nil {
		t.Error(err)
	}
	repository := repository.New(lruCache)
	evaluator, err := evaluation.NewEvaluator(repository)
	if err != nil {
		t.Error(err)
	}
	for _, useCase := range data {
		for identifier, value := range useCase.Expected {
			tests = append(tests, args{
				file:     useCase.Filename,
				target:   identifier,
				value:    value,
				testFile: useCase,
			})
		}
	}

	for _, testCase := range tests {
		t.Run(testCase.testFile.Filename, func(t *testing.T) {
			var target *rest.Target
			if testCase.target != "_no_target" {
				for i, val := range testCase.testFile.Targets {
					if val.Identifier == testCase.target {
						target = &testCase.testFile.Targets[i]
					}
				}
			}
			evaluator.BoolVariation(testCase.testFile.Flag.Feature, target, false)
		})
	}
}

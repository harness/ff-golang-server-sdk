package tests

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/harness/ff-golang-server-sdk/evaluation"

	"github.com/harness/ff-golang-server-sdk/log"
	"github.com/harness/ff-golang-server-sdk/pkg/repository"
	"github.com/harness/ff-golang-server-sdk/rest"
)

const source = "./ff-test-cases/tests"

type testFile struct {
	Filename string
	Flag     rest.FeatureConfig     `json:"flag"`
	Segments []rest.Segment         `json:"segments"`
	Targets  []evaluation.Target    `json:"targets"`
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
		expected interface{}
		testFile testFile
	}

	tests := make([]args, 0)
	data := loadFiles()
	lruCache, err := repository.NewLruCache(1000)
	if err != nil {
		t.Error(err)
	}
	repo := repository.New(lruCache)
	evaluator, err := evaluation.NewEvaluator(repo, nil)
	if err != nil {
		t.Error(err)
	}
	for _, useCase := range data {
		useCase.Flag.Feature += useCase.Filename
		repo.SetFlag(useCase.Flag)
		for _, segment := range useCase.Segments {
			repo.SetSegment(segment)
		}
		for identifier, value := range useCase.Expected {
			tests = append(tests, args{
				file:     useCase.Filename,
				target:   identifier,
				expected: value,
				testFile: useCase,
			})
		}
	}

	for _, testCase := range tests {
		t.Run(testCase.testFile.Filename, func(t *testing.T) {
			var target *evaluation.Target
			if testCase.target != "_no_target" {
				for i, val := range testCase.testFile.Targets {
					if val.Identifier == testCase.target {
						target = &testCase.testFile.Targets[i]
					}
				}
			}
			var got interface{}
			switch testCase.testFile.Flag.Kind {
			case "boolean":
				got = evaluator.BoolVariation(testCase.testFile.Flag.Feature, target, false)
			case "string":
				got = evaluator.StringVariation(testCase.testFile.Flag.Feature, target, "blue")
			case "int":
				got = evaluator.IntVariation(testCase.testFile.Flag.Feature, target, 100)
			case "number":
				got = evaluator.NumberVariation(testCase.testFile.Flag.Feature, target, 50.00)
			case "json":
				got = evaluator.JSONVariation(testCase.testFile.Flag.Feature, target, map[string]interface{}{})
			}
			if !reflect.DeepEqual(got, testCase.expected) {
				t.Errorf("eval engine got = %v, want %v", got, testCase.expected)
			}
		})
	}
}

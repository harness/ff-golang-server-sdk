package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/harness/ff-golang-server-sdk/logger"

	"github.com/harness/ff-golang-server-sdk/evaluation"

	"github.com/harness/ff-golang-server-sdk/log"
	"github.com/harness/ff-golang-server-sdk/pkg/repository"
	"github.com/harness/ff-golang-server-sdk/rest"
)

const source = "./ff-test-cases/tests"

type test struct {
	Flag     string      `json:"flag"`
	Target   *string     `json:"target"`
	Expected interface{} `json:"expected"`
}

type testFile struct {
	Filename string
	Flags    []rest.FeatureConfig `json:"flags"`
	Segments []rest.Segment       `json:"segments"`
	Targets  []evaluation.Target  `json:"targets"`
	Tests    []test               `json:"tests"`
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
	t.Parallel()
	fixtures := loadFiles()
	for _, fixture := range fixtures {
		lruCache, err := repository.NewLruCache(1000)
		if err != nil {
			t.Error(err)
		}
		repo := repository.New(lruCache)
		evaluator, err := evaluation.NewEvaluator(repo, nil, logger.NewNoOpLogger())
		if err != nil {
			t.Error(err)
		}
		for _, flag := range fixture.Flags {
			repo.SetFlag(flag)
		}
		for _, segment := range fixture.Segments {
			repo.SetSegment(segment)
		}

		for _, testCase := range fixture.Tests {
			testName := fmt.Sprintf("test fixture %s with flag %s", fixture.Filename, testCase.Flag)
			if testCase.Target != nil {
				testName = fmt.Sprintf("%s and target %s", testName, *testCase.Target)
			}
			t.Run(testName, func(t *testing.T) {
				var target *evaluation.Target
				if testCase.Target != nil {
					for i, val := range fixture.Targets {
						if val.Identifier == *testCase.Target {
							target = &fixture.Targets[i]
						}
					}
				}
				var got interface{}
				flag, err := repo.GetFlag(testCase.Flag)
				if err != nil {
					t.Errorf("flag %s not found", testCase.Flag)
				}
				switch flag.Kind {
				case rest.FeatureConfigKindBoolean:
					got = evaluator.BoolVariation(testCase.Flag, target, false)
				case rest.FeatureConfigKindString:
					got = evaluator.StringVariation(testCase.Flag, target, "blue")
				case rest.FeatureConfigKindInt:
					got = evaluator.IntVariation(testCase.Flag, target, 100)
				case "number":
					got = evaluator.NumberVariation(testCase.Flag, target, 50.00)
				case rest.FeatureConfigKindJson:
					got = evaluator.JSONVariation(testCase.Flag, target, map[string]interface{}{})
				}
				if !reflect.DeepEqual(got, testCase.Expected) {
					t.Errorf("eval engine got = %v, want %v", got, testCase.Expected)
				}
			})
		}
	}
}

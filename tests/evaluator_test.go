package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

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

	slice := []testFile{}
	err := filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(info.Name()) != ".json" {
				return nil
			}
			if f, err := loadFile(path); err == nil {
				slice = append(slice, f)
			} else {
				log.Errorf("unable to load %s because %v", info.Name(), err)
			}
			return nil
		})
	if err != nil {
		log.Error(err)
	}

	return slice
}

func loadFile(filename string) (testFile, error) {
	fp := filepath.Clean(filename)
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
			repo.SetFlag(flag, false)
		}
		for _, segment := range fixture.Segments {
			repo.SetSegment(segment, false)
		}

		for _, testCase := range fixture.Tests {
			testName := fixture.Filename
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
				case rest.FeatureConfigKindInt, "number":
					got = evaluator.NumberVariation(testCase.Flag, target, 50.00)
				case rest.FeatureConfigKindJson:
					got = evaluator.JSONVariation(testCase.Flag, target, map[string]interface{}{})
					str, _ := json.Marshal(&got)
					got = string(str)

				}

				if flag.Kind == rest.FeatureConfigKindJson {
					expected := fmt.Sprintf("%s", testCase.Expected)
					jsonStr := fmt.Sprintf("%s", got)
					assert.JSONEq(t, expected, jsonStr, "eval engine got = [%v], want [%v]", jsonStr, expected)
				} else {
					if !reflect.DeepEqual(got, testCase.Expected) {
						t.Errorf("eval engine got = [%v], want [%v]", got, testCase.Expected)
					}
				}

			})
		}
	}
}

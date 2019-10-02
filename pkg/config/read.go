package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/rancher/mapper"
	"github.com/rancher/mapper/convert"
	merge2 "github.com/rancher/mapper/convert/merge"
	"github.com/rancher/mapper/values"
	"github.com/sirupsen/logrus"
)

const (
	SystemConfig = "/k3os/system/config.yaml"
	LocalConfig  = "/var/lib/rancher/k3os/config.yaml"
	localConfigs = "/var/lib/rancher/k3os/config.d"
)

var (
	schemas = mapper.NewSchemas().Init(func(s *mapper.Schemas) *mapper.Schemas {
		s.DefaultMappers = func() []mapper.Mapper {
			return []mapper.Mapper{
				NewToSlice(),
				NewToBool(),
				&FuzzyNames{},
			}
		}
		return s
	}).MustImport(CloudConfig{})
	schema  = schemas.Schema("cloudConfig")
	readers = []reader{
		readSystemConfig,
		readCmdline,
		readLocalConfig,
		readCloudConfig,
		readUserData,
	}
)

func init() {
	fmt.Println("====> Fake MAIN =====")
	logrus.SetLevel(logrus.DebugLevel)
	fmt.Printf("schema: %#v", schema)
	mapper.YAMLEncoder(os.Stdout, schema)

	//spew.Dump(schema)
}

func ToEnv(cfg CloudConfig) ([]string, error) {
	data, err := convert.EncodeToMap(&cfg)
	if err != nil {
		return nil, err
	}

	return mapToEnv("", data), nil
}

func mapToEnv(prefix string, data map[string]interface{}) []string {
	var result []string
	for k, v := range data {
		keyName := strings.ToUpper(prefix + convert.ToYAMLKey(k))
		if data, ok := v.(map[string]interface{}); ok {
			subResult := mapToEnv(keyName+"_", data)
			result = append(result, subResult...)
		} else {
			result = append(result, fmt.Sprintf("%s=%v", keyName, v))
		}
	}
	return result
}

func ReadConfig() (CloudConfig, error) {
	return readersToObject(append(readers, readLocalConfigs()...)...)
}

func readersToObject(readers ...reader) (CloudConfig, error) {
	result := CloudConfig{
		K3OS: K3OS{
			Install: &Install{},
		},
	}

	data, err := merge(readers...)
	y, _ := yaml.Marshal(data)
	logrus.Debugf("HACKED MERGED!!! err:%#v data:\n===\n%s\n===\n", err, y)
	if err != nil {
		return result, err
	}

	return result, convert.ToObj(data, &result)
}

type reader func() (string, map[string]interface{}, error)

func merge(readers ...reader) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	for _, r := range readers {
		readername, newData, err := r()
		logrus.Debugf("HACK merge() next reader: %#v err:%v", readername, err)
		logrus.Debugf("HACK merge() newData: %#v", newData)
		if err != nil {
			return nil, err
		}

		if err := schema.Mapper.ToInternal(newData); err != nil {
			return nil, err
		}
		logrus.Debug("HACK merge2.UpdateMerge start ...")
		data = merge2.UpdateMerge(schema, schemas, data, newData, false)
		logrus.Debugf("HACK merge2.UpdateMerge END MERGED_DATA: %#v", data)
		//y, _ := yaml.Marshal(data)
		//logrus.Debugf("HACK merge() MERGED: \n====\n%s\n====", y)

	}
	return data, nil
}

func readSystemConfig() (string, map[string]interface{}, error) {
	return readFile(SystemConfig)
}

func readLocalConfig() (string, map[string]interface{}, error) {
	return readFile(LocalConfig)
}

func readLocalConfigs() []reader {
	var result []reader

	files, err := ioutil.ReadDir(localConfigs)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return []reader{
			func() (string, map[string]interface{}, error) {
				return "readLocalConfigs", nil, err
			},
		}
	}

	for _, f := range files {
		p := filepath.Join(localConfigs, f.Name())
		result = append(result, func() (string, map[string]interface{}, error) {
			return readFile(p)
		})
	}

	return result
}

func readFile(path string) (string, map[string]interface{}, error) {
	f, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return path, nil, nil
	} else if err != nil {
		return path, nil, err
	}

	data := map[string]interface{}{}
	if err := yaml.Unmarshal(f, &data); err != nil {
		return path, nil, err
	}

	return path, data, nil
}

func readCmdline() (string, map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile("/proc/cmdline")
	if os.IsNotExist(err) {
		return "readCmdline", nil, nil
	} else if err != nil {
		return "readCmdline", nil, err
	}

	data := map[string]interface{}{}
	for _, item := range strings.Fields(string(bytes)) {
		parts := strings.SplitN(item, "=", 2)
		value := "true"
		if len(parts) > 1 {
			value = parts[1]
		}
		keys := strings.Split(parts[0], ".")
		existing, ok := values.GetValue(data, keys...)
		if ok {
			switch v := existing.(type) {
			case string:
				values.PutValue(data, []string{v, value}, keys...)
			case []string:
				values.PutValue(data, append(v, value), keys...)
			}
		} else {
			values.PutValue(data, value, keys...)
		}
	}

	return "readCmdline", data, nil
}

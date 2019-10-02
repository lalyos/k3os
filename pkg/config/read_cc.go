package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/rancher/mapper/convert"
	"github.com/sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
)

const (
	hostname = "/run/config/local_hostname"
	ssh      = "/run/config/ssh/authorized_keys"
	userdata = "/run/config/userdata"
)

func readCloudConfig() (string, map[string]interface{}, error) {
	logrus.Debug("HACK readCloudConfig ")
	var keys []string
	result := map[string]interface{}{}

	hostname, err := ioutil.ReadFile(hostname)
	if err == nil {
		result["hostname"] = strings.TrimSpace(string(hostname))
	}

	keyData, err := ioutil.ReadFile(ssh)
	if err != nil {
		// ignore error
		return "readCloudConfig", result, nil
	}

	for _, line := range strings.Split(string(keyData), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			keys = append(keys, line)
		}
	}

	if len(keys) > 0 {
		result["ssh_authorized_keys"] = keys
	}

	return "readCloudConfig", result, nil
}

func readUserData() (string, map[string]interface{}, error) {
	logrus.Debug("HACK UserData")
	result := map[string]interface{}{}

	data, err := ioutil.ReadFile(userdata)
	if os.IsNotExist(err) {
		return "readUserData", nil, nil
	} else if err != nil {
		return "readUserData", nil, err
	}

	cc := CloudConfig{}
	script := false
	if bytes.Contains(data, []byte{0}) {
		script = true
		cc.WriteFiles = []File{
			{
				Content:  base64.StdEncoding.EncodeToString(data),
				Encoding: "b64",
			},
		}
		cc.Runcmd = []string{"/run/k3os/userdata"}
	} else if strings.HasPrefix(string(data), "#!") {
		script = true
		cc.WriteFiles = []File{
			{
				Content: string(data),
			},
		}
	}

	if script {
		cc.WriteFiles[0].Owner = "root"
		cc.WriteFiles[0].RawFilePermissions = "0700"
		cc.WriteFiles[0].Path = "/run/k3os/userdata"
		cc.Runcmd = []string{"/run/k3os/userdata"}

		m, err := convert.EncodeToMap(cc)
		return "readUserData", m, err
	}

	yaml.Unmarshal(data, &result)
	logrus.Debug("HACK ", fmt.Sprintf("data: %#v", result))

	return "readUserData",z result, yaml.Unmarshal(data, &result)
}

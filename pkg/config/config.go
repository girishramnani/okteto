// Copyright 2020 The Okteto Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/okteto/okteto/pkg/log"
	"github.com/okteto/okteto/pkg/model"
)

const (
	oktetoFolderName = ".okteto"
)

// VersionString the version of the cli
var VersionString string

var timeout time.Duration
var tOnce sync.Once

//GetBinaryName returns the name of the binary
func GetBinaryName() string {
	return filepath.Base(GetBinaryFullPath())
}

//GetBinaryFullPath returns the name of the binary
func GetBinaryFullPath() string {
	return os.Args[0]
}

// GetOktetoHome returns the path of the okteto folder
func GetOktetoHome() string {
	if v, ok := os.LookupEnv("OKTETO_FOLDER"); ok {
		if !model.FileExists(v) {
			log.Fatalf("OKTETO_FOLDER doesn't exist: %s", v)
		}

		return v
	}

	home := GetUserHomeDir()
	d := filepath.Join(home, oktetoFolderName)

	if err := os.MkdirAll(d, 0700); err != nil {
		log.Fatalf("failed to create %s: %s", d, err)
	}

	return d
}

// GetNamespaceHome returns the path of the folder
func GetNamespaceHome(namespace string) string {
	okHome := GetOktetoHome()
	d := filepath.Join(okHome, namespace)

	if err := os.MkdirAll(d, 0700); err != nil {
		log.Fatalf("failed to create %s: %s", d, err)
	}

	return d
}

// GetDeploymentHome returns the path of the folder
func GetDeploymentHome(namespace, name string) string {
	okHome := GetOktetoHome()
	d := filepath.Join(okHome, namespace, name)

	if err := os.MkdirAll(d, 0700); err != nil {
		log.Fatalf("failed to create %s: %s", d, err)
	}

	return d
}

// GetUserHomeDir returns the OS home dir
func GetUserHomeDir() string {
	if v, ok := os.LookupEnv("OKTETO_HOME"); ok {
		if !model.FileExists(v) {
			log.Fatalf("OKTETO_HOME points to a non-existing directory: %s", v)
		}

		return v
	}

	if runtime.GOOS == "windows" {
		home, err := homedirWindows()
		if err != nil {
			log.Fatalf("couldn't determine your home directory: %s", err)
		}

		return home
	}

	return os.Getenv("HOME")

}

func homedirWindows() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	if home := os.Getenv("USERPROFILE"); home != "" {
		return home, nil
	}

	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		return "", fmt.Errorf("HOME, HOMEDRIVE, HOMEPATH, or USERPROFILE are empty. Use $OKTETO_HOME to set your home directory")
	}

	return home, nil
}

// GetKubeConfigFile returns the path to the kubeconfig file, taking the KUBECONFIG env var into consideration
func GetKubeConfigFile() string {
	home := GetUserHomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if len(kubeconfigEnv) > 0 {
		kubeconfig = splitKubeConfigEnv(kubeconfigEnv)
	}
	return kubeconfig
}

func splitKubeConfigEnv(value string) string {
	if runtime.GOOS == "windows" {
		return strings.Split(value, ";")[0]
	}
	return strings.Split(value, ":")[0]
}

// GetTimeout returns the per-action timeout
func GetTimeout() time.Duration {
	tOnce.Do(func() {
		timeout = (30 * time.Second)
		t, ok := os.LookupEnv("OKTETO_TIMEOUT")
		if !ok {
			return
		}

		parsed, err := time.ParseDuration(t)
		if err != nil {
			log.Infof("'%s' is not a valid duration, ignoring", t)
			return
		}

		log.Infof("OKTETO_TIMEOUT applied: '%s'", parsed.String())
		timeout = parsed
	})

	return timeout
}

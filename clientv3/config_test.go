// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clientv3

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
)

var (
	certPath       = "../integration/fixtures/server.crt"
	privateKeyPath = "../integration/fixtures/server.key.insecure"
	caPath         = "../integration/fixtures/ca.crt"
)

func TestConfigFromFile(t *testing.T) {
	tests := []struct {
		ym *YamlConfig

		werr bool
	}{
		{
			&YamlConfig{},
			false,
		},
		{
			&YamlConfig{
				InsecureTransport: true,
			},
			false,
		},
		{
			&YamlConfig{
				Keyfile:               privateKeyPath,
				Certfile:              certPath,
				CAfile:                caPath,
				InsecureSkipTLSVerify: true,
			},
			false,
		},
		{
			&YamlConfig{
				Keyfile:  "bad",
				Certfile: "bad",
			},
			true,
		},
		{
			&YamlConfig{
				Keyfile:  privateKeyPath,
				Certfile: certPath,
				CAfile:   "bad",
			},
			true,
		},
	}

	for i, tt := range tests {
		tmpfile, err := ioutil.TempFile("", "clientcfg")
		if err != nil {
			log.Fatal(err)
		}

		b, err := yaml.Marshal(tt.ym)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := tmpfile.Write(b); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}

		cfg, err := configFromFile(tmpfile.Name())
		if err != nil && !tt.werr {
			t.Errorf("#%d: err = %v, want %v", i, err, tt.werr)
			continue
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(cfg.Endpoints, tt.ym.Endpoints) {
			t.Errorf("#%d: endpoint = %v, want %v", i, cfg.Endpoints, tt.ym.Endpoints)
		}

		if tt.ym.InsecureTransport != (cfg.TLS == nil) {
			t.Errorf("#%d: insecureTransport = %v, want %v", i, cfg.TLS == nil, tt.ym.InsecureTransport)
		}

		if !tt.ym.InsecureTransport {
			if tt.ym.Certfile != "" && len(cfg.TLS.Certificates) == 0 {
				t.Errorf("#%d: failed to load in cert", i)
			}
			if tt.ym.CAfile != "" && cfg.TLS.RootCAs == nil {
				t.Errorf("#%d: failed to load in ca cert", i)
			}
			if cfg.TLS.InsecureSkipVerify != tt.ym.InsecureSkipTLSVerify {
				t.Errorf("#%d: skipTLSVeify = %v, want %v", cfg.TLS.InsecureSkipVerify, tt.ym.InsecureSkipTLSVerify)
			}
		}

		os.Remove(tmpfile.Name())
	}
}

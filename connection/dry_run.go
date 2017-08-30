// Copyright © 2017 ben dewan <benj.dewan@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package connection

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"

	"github.com/ghodss/yaml"
)

var schemes = map[string]string{
	"postgresql":    "postgres",
	"redis":         "redis",
	"rabbitmq":      "amqps",
	"elasticsearch": "https",
	"etcd":          "https",
	"janus":         "https",
	"scylla":        "https",
	"mongodb":       "mongodb",
	"mysql":         "mysql",
}

var userpass = [][]string{
	{"mario", "175Am31"},
	{"luigi", "gr33nM4r10"},
	{"zelda", "br347h0ft3hw1ld"},
	{"gfreeman", "1,2..."},
	{"doomguy", "s3cr37_r00m5"},
	{"bjblazkowicz", "h3lm3t5t4ck5"},
	{"admin", "admin"},
	{"alice", "Ez57510qVFnK7obJYKr3"},
}

func dryRunCreate(cxn *Connection, deployment Deployment) error {
	cxn.newDeploymentIDs.Store(fakeID(deployment), struct{}{})
	return nil
}

func dryRunUpdate(cxn *Connection, deployment Deployment) error {
	existing, ok := cxn.getDeploymentByName(deployment.GetName())
	if !ok {
		return fmt.Errorf("Attempting to update '%s', but it doesn't exist",
			deployment.GetName())
	}
	cxn.newDeploymentIDs.Store(existing.ID, struct{}{})
	return nil
}

func fakeOutputYAML(id string) ([]byte, error) {
	cxnYAML := make(map[string]outputYAML)
	segments := strings.Split(id, "::")
	cxnYAML[segments[1]] = outputYAML{
		Type:        segments[0],
		CACert:      fakeCA(),
		Connections: fakeConnectionYAML(segments[0], segments[1]),
	}
	return yaml.Marshal(cxnYAML)
}

func fakeConnectionYAML(deployType, deployName string) []connectionYAML {
	rando := userpass[rand.Intn(len(userpass))]
	return []connectionYAML{
		{
			Scheme:   schemes[deployType],
			Host:     "pachelbel-dry-run.compose.direct",
			Port:     rand.Intn(6497 + 3),
			Path:     fakePath(deployType, deployName),
			Username: rando[0],
			Password: rando[1],
		},
	}
}

func fakePath(deployType, deployName string) string {
	switch deployType {
	case "postgresql":
		return "/compose"
	case "rabbitmq":
		return "/" + deployName
	case "elasticSearch":
		return "/"
	default:
		return ""
	}
}

func fakeCA() string {
	var buf bytes.Buffer
	buf.WriteString("-----BEGIN CERTIFICATE-----\n")
	data := make([]byte, 1024)
	rand.Read(data)
	encData := []byte(base64.StdEncoding.EncodeToString(data))
	for i := 64; i < len(encData); i += 65 {
		encData = append(encData[:i], append([]byte("\n"), encData[i:]...)...)
	}
	buf.Write(encData)
	buf.WriteString("\n-----END CERTIFICATE-----\n")
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func fakeID(deployment Deployment) string {
	return fmt.Sprintf("%s::%s", deployment.GetType(),
		deployment.GetName())
}

func isFake(id string) bool {
	return strings.Contains(id, "::")
}

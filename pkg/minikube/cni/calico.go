/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cni

import (
	"bytes"
	// goembed needs this
	_ "embed"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
)

// https://docs.projectcalico.org/manifests/calico.yaml
//
//go:embed calico.yaml
var calicoYaml string

// calicoTmpl is from https://docs.projectcalico.org/manifests/calico.yaml
var calicoTmpl = template.Must(template.New("calico").Parse(calicoYaml))

// Calico is the Calico CNI manager
type Calico struct {
	cc config.ClusterConfig
}

type calicoTmplStruct struct {
	DeploymentImageName  string
	DaemonSetImageName   string
	FelixDriverImageName string
	BinaryImageName      string
}

// String returns a string representation of this CNI
func (c Calico) String() string {
	return "Calico"
}

// manifest returns a Kubernetes manifest for a CNI
func (c Calico) manifest() (assets.CopyableFile, error) {
	input := &calicoTmplStruct{
		DeploymentImageName:  images.CalicoDeployment(c.cc.KubernetesConfig.ImageRepository),
		DaemonSetImageName:   images.CalicoDaemonSet(c.cc.KubernetesConfig.ImageRepository),
		FelixDriverImageName: images.CalicoFelixDriver(c.cc.KubernetesConfig.ImageRepository),
		BinaryImageName:      images.CalicoBin(c.cc.KubernetesConfig.ImageRepository),
	}

	b := bytes.Buffer{}
	if err := calicoTmpl.Execute(&b, input); err != nil {
		return nil, err
	}
	return manifestAsset(b.Bytes()), nil
}

// Apply enables the CNI
func (c Calico) Apply(r Runner) error {
	m, err := c.manifest()
	if err != nil {
		return errors.Wrap(err, "manifest")
	}
	return applyManifest(c.cc, r, m)
}

// CIDR returns the default CIDR used by this CNI
func (c Calico) CIDR() string {
	// Calico docs specify 192.168.0.0/16 - but we do this for compatibility with other CNI's.
	return DefaultPodCIDR
}

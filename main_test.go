package main

import (
	"reflect"
	"testing"

	"k8s.io/client-go/tools/clientcmd/api/v1"
)

func Test_readFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"empty",
			args{""},
			"",
			false,
		},
		{
			"error no file",
			args{"testdata/readFile.none"},
			"",
			true,
		},
		{
			"empty file",
			args{"testdata/readFile.empty"},
			"",
			false,
		},
		{
			"file having content foo",
			args{"testdata/readFile.foo"},
			"foo",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("readFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readKubeconfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.Config
		wantErr bool
	}{
		{
			"empty",
			args{"testdata/readKubeconfig.empty"},
			nil,
			false,
		},
		{
			"simple",
			args{"testdata/readKubeconfig.simple"},
			&v1.Config{
				APIVersion:     "v1",
				Kind:           "Config",
				CurrentContext: "kubernetes-admin@kubernetes",
			},
			false,
		},
		{
			"invalid yaml",
			args{"testdata/readKubeconfig.invaldyaml"},
			nil,
			true,
		},
		{
			"no file",
			args{"testdata/readKubeconfig.none"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readKubeconfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("readKubeconfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readKubeconfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeKubeconfig(t *testing.T) {
	type args struct {
		sourcePath      string
		destinationPath string
		name            string
		force           bool
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.Config
		wantErr bool
	}{
		{
			"empty empty",
			args{"testdata/mergeKubeconfig.empty", "testdata/mergeKubeconfig.empty", "", false},
			nil,
			true,
		},
		{
			"noContext empty",
			args{"testdata/mergeKubeconfig.noContext", "testdata/mergeKubeconfig.empty", "", false},
			nil,
			true,
		},
		{
			"none empty",
			args{"testdata/mergeKubeconfig.none", "testdata/mergeKubeconfig.empty", "", false},
			nil,
			true,
		},
		{
			"noContext none",
			args{"testdata/mergeKubeconfig.noContext", "testdata/mergeKubeconfig.none", "", false},
			nil,
			true,
		},
		{
			"onlyContext empty",
			args{"testdata/mergeKubeconfig.onlyContext", "testdata/mergeKubeconfig.empty", "", false},
			nil,
			true,
		},
		{
			"noCluster empty",
			args{"testdata/mergeKubeconfig.noCluster", "testdata/mergeKubeconfig.empty", "", false},
			nil,
			true,
		},
		{
			"complete empty",
			args{"testdata/mergeKubeconfig.complete", "testdata/mergeKubeconfig.empty", "", false},
			&v1.Config{
				APIVersion:     "v1",
				Kind:           "Config",
				CurrentContext: "kubernetes-admin@kubernetes",
				Contexts: []v1.NamedContext{{
					Name: "kubernetes-admin@kubernetes",
					Context: v1.Context{
						Cluster:  "kubernetes",
						AuthInfo: "kubernetes-admin",
					},
				}},
				Clusters: []v1.NamedCluster{{
					Name: "kubernetes",
					Cluster: v1.Cluster{
						Server: "https://foo.bar:6443",
						CertificateAuthorityData: []byte("Foo"),
					},
				}},
				AuthInfos: []v1.NamedAuthInfo{{
					Name: "kubernetes-admin",
					AuthInfo: v1.AuthInfo{
						ClientCertificateData: []byte("Bar"),
						ClientKeyData:         []byte("Foobar"),
					},
				}},
			},
			false,
		},
		{
			"complete complete",
			args{"testdata/mergeKubeconfig.complete", "testdata/mergeKubeconfig.complete", "", false},
			&v1.Config{
				APIVersion:     "v1",
				Kind:           "Config",
				CurrentContext: "kubernetes-admin@kubernetes",
				Contexts: []v1.NamedContext{{
					Name: "kubernetes-admin@kubernetes",
					Context: v1.Context{
						Cluster:  "kubernetes",
						AuthInfo: "kubernetes-admin",
					},
				}},
				Clusters: []v1.NamedCluster{{
					Name: "kubernetes",
					Cluster: v1.Cluster{
						Server: "https://foo.bar:6443",
						CertificateAuthorityData: []byte("Foo"),
					},
				}},
				AuthInfos: []v1.NamedAuthInfo{{
					Name: "kubernetes-admin",
					AuthInfo: v1.AuthInfo{
						ClientCertificateData: []byte("Bar"),
						ClientKeyData:         []byte("Foobar"),
					},
				}},
			},
			false,
		},
		{
			"complete complete force",
			args{"testdata/mergeKubeconfig.complete", "testdata/mergeKubeconfig.complete", "", true},
			&v1.Config{
				APIVersion:     "v1",
				Kind:           "Config",
				CurrentContext: "kubernetes-admin@kubernetes",
				Contexts: []v1.NamedContext{{
					Name: "kubernetes-admin@kubernetes",
					Context: v1.Context{
						Cluster:  "kubernetes",
						AuthInfo: "kubernetes-admin",
					},
				}},
				Clusters: []v1.NamedCluster{{
					Name: "kubernetes",
					Cluster: v1.Cluster{
						Server: "https://foo.bar:6443",
						CertificateAuthorityData: []byte("Foo"),
					},
				}},
				AuthInfos: []v1.NamedAuthInfo{{
					Name: "kubernetes-admin",
					AuthInfo: v1.AuthInfo{
						ClientCertificateData: []byte("Bar"),
						ClientKeyData:         []byte("Foobar"),
					},
				}},
			},
			false,
		},
		{
			"complete empty .name=foo",
			args{"testdata/mergeKubeconfig.complete", "testdata/mergeKubeconfig.empty", "foo", false},
			&v1.Config{
				APIVersion:     "v1",
				Kind:           "Config",
				CurrentContext: "foo",
				Contexts: []v1.NamedContext{{
					Name: "foo",
					Context: v1.Context{
						Cluster:  "foo-kubernetes",
						AuthInfo: "foo-kubernetes-admin",
					},
				}},
				Clusters: []v1.NamedCluster{{
					Name: "foo-kubernetes",
					Cluster: v1.Cluster{
						Server: "https://foo.bar:6443",
						CertificateAuthorityData: []byte("Foo"),
					},
				}},
				AuthInfos: []v1.NamedAuthInfo{{
					Name: "foo-kubernetes-admin",
					AuthInfo: v1.AuthInfo{
						ClientCertificateData: []byte("Bar"),
						ClientKeyData:         []byte("Foobar"),
					},
				}},
			},
			false,
		},
	}
	for _, tt := range tests {
		name = tt.args.name
		force = tt.args.force
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergeKubeconfig(tt.args.sourcePath, tt.args.destinationPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeKubeconfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeKubeconfig() = %v, want %v", got, tt.want)
			}
		})
	}

	// execute tests having a set name
	name = "foo"
}

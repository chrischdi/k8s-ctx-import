package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/ghodss/yaml"

	"k8s.io/client-go/tools/clientcmd/api/v1"
)

// The usage of a manually defined `FlagSet`` hides the flags created by the
// `glog`` library which gets imported somewhere in `client-go`.
var (
	force bool
	name  string
	help  bool
	flags = new(flag.FlagSet)
	Usage = func() {
		fmt.Fprintf(os.Stderr, "`k8s-ctx-import` is an utility to merge kubernetes contexts to a single kubeconfig.\n")
		fmt.Fprintf(os.Stderr, "It imports the context either to `~/.kube/config` or to the file defined by the `KUBECONFIG` environment variable.\n")
		fmt.Fprintf(os.Stderr, "Usage of k8s-ctx-import:\n")
		flags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "    cat /some/kubeconfig | k8s-ctx-import\n")
	}
)

func init() {
	flag.Usage = Usage
	flags.BoolVar(&help, "help", false, "display this help and exit")
	flags.BoolVar(&force, "force", false, "force import of context")
	flags.StringVar(&name, "name", "", "renames the context for the import")
}

// reads a configfile from `path`. If `path == ""` it will read from stdin.
func readConfig(path string) (*v1.Config, error) {
	var b []byte
	var err error
	if path == "" {
		b, err = ioutil.ReadAll(os.Stdin)
	} else {
		b, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	c := &v1.Config{}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func main() {
	flag.Parse()

	conf, err := readConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read stdin: %v\n", err)
	}

	// determine output file
	var gconfFile string
	if os.Getenv("KUBECONFIG") != "" {
		gconfFile = os.Getenv("KUBECONFIG")
	} else {
		gconfFile = path.Join(os.Getenv("HOME"), ".kube", "config")
	}
	gconf, err := readConfig(gconfFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: unable to read file %s: %v\n", gconfFile, err)
	}
	if gconf == nil {
		gconf = &v1.Config{
			APIVersion: "v1",
			Kind:       "Config",
		}
	}

	var ctx *v1.NamedContext
	for _, c := range conf.Contexts {
		if c.Name == conf.CurrentContext {
			ctx = &c
		}
	}
	if ctx == nil {
		fmt.Fprintf(os.Stderr, "ERROR: context %s not found\n", conf.CurrentContext)
		os.Exit(1)
	}

	var authInfo *v1.NamedAuthInfo
	for _, a := range conf.AuthInfos {
		if a.Name == ctx.Context.AuthInfo {
			authInfo = &a
		}
	}
	if authInfo == nil {
		fmt.Fprintf(os.Stderr, "ERROR: authInfo %s not found\n", ctx.Context.AuthInfo)
		os.Exit(1)
	}

	var cluster *v1.NamedCluster
	for _, c := range conf.Clusters {
		if c.Name == ctx.Context.Cluster {
			cluster = &c
		}
	}
	if cluster == nil {
		fmt.Fprintf(os.Stderr, "ERROR: cluster %s not found\n", ctx.Context.Cluster)
		os.Exit(1)
	}

	if name != "" {
		ctx.Name = name
		ctx.Context.Cluster = name + "-" + cluster.Name
		ctx.Context.AuthInfo = name + "-" + authInfo.Name
		cluster.Name = name + "-" + cluster.Name
		authInfo.Name = name + "-" + authInfo.Name
	}

	var exists bool
	// check if context having the same name already exists
	exists = false
	for _, c := range gconf.Contexts {
		if c.Name == ctx.Name {
			if force {
				c.Context = ctx.Context
			} else {
				fmt.Fprintf(os.Stderr, "WARN: context having the same name (%s) already exists\n", c.Name)
			}
			exists = true
		}
	}
	if !exists {
		gconf.Contexts = append(gconf.Contexts, *ctx)
	}

	// check if cluster having the same name already exists
	exists = false
	for _, c := range gconf.Clusters {
		if c.Name == cluster.Name {
			if force {
				c.Cluster = cluster.Cluster
			} else {
				fmt.Fprintf(os.Stderr, "WARN: cluster information having the same name (%s) already exists\n", c.Name)
			}
			exists = true
		}
	}
	if !exists {
		gconf.Clusters = append(gconf.Clusters, *cluster)
	}

	// check if authInfo having the same name already exists
	exists = false
	for _, a := range gconf.AuthInfos {
		if a.Name == authInfo.Name {
			if force {
				a.AuthInfo = authInfo.AuthInfo
			} else {
				fmt.Fprintf(os.Stderr, "WARN: authentication information having the same name (%s) already exists\n", a.Name)
			}
			exists = true
		}
	}
	if !exists {
		gconf.AuthInfos = append(gconf.AuthInfos, *authInfo)
	}

	gconf.CurrentContext = ctx.Name

	b, err := yaml.Marshal(gconf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: error Marshaling new kubeconfig\n")
	}

	err = ioutil.WriteFile(gconfFile, b, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to write file %s\n", gconfFile)
	}
}

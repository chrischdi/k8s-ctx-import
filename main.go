package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

var (
	force             bool
	setCurrentContext bool
	name              string
	stdout            bool
	help              bool
	flags             = new(flag.FlagSet)
	Usage             = func() {
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
	flags.BoolVar(&help, "h", false, "")
	flags.BoolVar(&help, "help", false, "display this help and exit")
	flags.BoolVar(&force, "force", false, "force import of context")
	flags.BoolVar(&setCurrentContext, "set-current-context", true, "set current context to imported context")
	flags.StringVar(&name, "name", "", "renames the context for the import")
	flags.BoolVar(&stdout, "stdout", false, "print result to stdout instead of writing to file")
}

// readFile reads from stdin (if path is empty) or from a file and returns its string
func readFile(path string) ([]byte, error) {
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
	return b, err
}

func readKubeconfig(path string) (*v1.Config, error) {
	d := &v1.Config{}
	c, err := readFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(c, &d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func main() {
	flags.Parse(os.Args[1:])
	if help {
		Usage()
		os.Exit(0)
	}

	// determine output file
	destinationPath := os.Getenv("KUBECONFIG")
	if destinationPath == "" {
		destinationPath = path.Join(os.Getenv("HOME"), ".kube", "config")
	}

	// set output to stderr
	log.SetOutput(os.Stderr)

	// "" means we read from stdin
	newcfg, err := mergeKubeconfig("", destinationPath)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	b, err := yaml.Marshal(newcfg)
	if err != nil {
		log.Fatalf("error Marshaling new kubeconfig\n")
	}

	if destinationPath != "" && !stdout {
		err = ioutil.WriteFile(destinationPath, b, 0644)
		if err != nil {
			log.Fatalf("unable to write file %s\n", destinationPath)
		}
	} else {
		fmt.Println(string(b))
	}
}

// mergeKubeconfig tries to merge the kubeconfig at sourcePath (stdin if empty string)
// to the kubeconfig at destinationPath and returns the result
func mergeKubeconfig(sourcePath, destinationPath string) (*v1.Config, error) {
	source, err := readKubeconfig(sourcePath)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, fmt.Errorf("source kubeconfig is empty")
	}

	destination, err := readKubeconfig(destinationPath)
	if err != nil {
		return nil, err
	}

	if destination == nil {
		destination = &v1.Config{
			APIVersion: "v1",
			Kind:       "Config",
		}
	}

	// extract context, auth and cluster information from source

	var ctx *v1.NamedContext
	for _, c := range source.Contexts {
		if c.Name == source.CurrentContext {
			ctx = &c
			break
		}
	}
	if ctx == nil {
		return nil, fmt.Errorf("ERROR: context %s not found", source.CurrentContext)
	}

	var authInfo *v1.NamedAuthInfo
	for _, a := range source.AuthInfos {
		if a.Name == ctx.Context.AuthInfo {
			authInfo = &a
			break
		}
	}
	if authInfo == nil {
		return nil, fmt.Errorf("authInfo %s not found", ctx.Context.AuthInfo)
	}

	var cluster *v1.NamedCluster
	for _, c := range source.Clusters {
		if c.Name == ctx.Context.Cluster {
			cluster = &c
			break
		}
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster %s not found", ctx.Context.Cluster)
	}

	// set new context name if flag is set
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
	for i, c := range destination.Contexts {
		if c.Name == ctx.Name {
			if force {
				destination.Contexts[i] = *ctx
			} else {
				log.Printf("WARN: context having the same name (%s) already exists\n", c.Name)
			}
			exists = true
			break
		}
	}
	if !exists {
		destination.Contexts = append(destination.Contexts, *ctx)
	}

	// check if cluster having the same name already exists
	exists = false
	for i, c := range destination.Clusters {
		if c.Name == cluster.Name {
			if force {
				destination.Clusters[i] = *cluster
			} else {
				log.Printf("WARN: cluster information having the same name (%s) already exists\n", c.Name)
			}
			exists = true
			break
		}
	}
	if !exists {
		destination.Clusters = append(destination.Clusters, *cluster)
	}

	// check if authInfo having the same name already exists
	exists = false
	for i, a := range destination.AuthInfos {
		if a.Name == authInfo.Name {
			if force {
				destination.AuthInfos[i] = *authInfo
			} else {
				log.Printf("WARN: authentication information having the same name (%s) already exists\n", a.Name)
			}
			exists = true
			break
		}
	}
	if !exists {
		destination.AuthInfos = append(destination.AuthInfos, *authInfo)
	}

	if setCurrentContext {
		destination.CurrentContext = ctx.Name
	}

	return destination, nil
}

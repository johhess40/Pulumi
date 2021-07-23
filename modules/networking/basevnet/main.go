/*
NOTE: This is a first go at creating reusable modules using Pulumi
NOTE: Developers should work primarily with the code for the modules
NOTE: This code should leverage YAML or JSON for those who aren't devs
	  -This will allow users across an org to use the same modules without worrying about code
	  -YAML should be preferred, but JSON is sort of acceptable...
*/

package main

import (
	network "github.com/pulumi/pulumi-azure-native/sdk/go/azure/network"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type VirtNets struct {
	Rgs []struct {
		Rg struct {
			Friendly string `yaml:"friendly"`
			Name     string `yaml:"name"`
			Location string `yaml:"location"`
		} `yaml:"rg"`
	} `yaml:"rgs"`
	Networks []struct {
		Net struct {
			Friendly string `yaml:"friendly"`
			Name     string `yaml:"name"`
			Addspace string `yaml:"addspace"`
			Rgname   string `yaml:"rgname"`
			Location string `yaml:"location"`
		} `yaml:"net"`
	} `yaml:"networks"`
}

var doc string

func main() {
	readDir()
	vNets(doc)
}

//func vNets takes output of readFile and uses it to create resource groups and virtual networks
func vNets(net string) {

	nets, err := readFile(net)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%T\n", nets.Rgs)

	pulumi.Run(func(ctx *pulumi.Context) error {
		for _, v := range nets.Rgs {

			_, err := resources.NewResourceGroup(ctx, v.Rg.Friendly, &resources.ResourceGroupArgs{
				Location:          pulumi.String(v.Rg.Location),
				ResourceGroupName: pulumi.String(v.Rg.Name),
			})

			if err != nil {
				log.Fatalf("Resource Group creation failed with error: %v", err)
			}
		}

		time.Sleep(time.Duration(10) * time.Second)

		for _, v := range nets.Networks {
			_, err := network.NewVirtualNetwork(ctx, v.Net.Friendly, &network.VirtualNetworkArgs{
				AddressSpace: &network.AddressSpaceArgs{
					AddressPrefixes: pulumi.StringArray{
						pulumi.String(v.Net.Addspace),
					},
				},
				Location:           pulumi.String(v.Net.Location),
				ResourceGroupName:  pulumi.String(v.Net.Rgname),
				VirtualNetworkName: pulumi.String(v.Net.Name),
			})
			if err != nil {
				log.Fatalf("VNET deployment failed with error: %v", err)
			}

		}
		return nil
	})
}

//func readFile reads our YAML file and creates a map of string to interface
func readFile(x string) (*VirtNets, error) {
	file, err := ioutil.ReadFile(x)
	if err != nil {
		log.Fatalf("ioutil.ReadFile failed with error %v", err)
	}

	configs := VirtNets{}

	err = yaml.Unmarshal(file, &configs)
	if err != nil {
		log.Fatalf("Error in file %v", err)
	}

	return &configs, nil
}

//func readDir reads the directory for files that have a .yml or .yaml file extension
func readDir() {
	err := filepath.Walk("./conf", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var files string = path

		getconf := strings.Contains(files, filepath.Ext(".yml")) || strings.Contains(files, filepath.Ext(".yaml"))

		if getconf {
			doc = files
		}
		return nil

	})
	if err != nil {
		log.Fatalf("Unable to parse directory for conf folder, see error: %v", err)
	}
}

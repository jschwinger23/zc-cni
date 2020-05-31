package main

import (
	"context"
	"fmt"
	"os"

	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	"github.com/projectcalico/libcalico-go/lib/clientv3"
	calicoipam "github.com/projectcalico/libcalico-go/lib/ipam"
	caliconet "github.com/projectcalico/libcalico-go/lib/net"
	"github.com/projectcalico/libcalico-go/lib/options"
	osutils "github.com/projectcalico/libnetwork-plugin/utils/os"
	"github.com/prometheus/common/log"
)

func main() {
	if len(os.Args) != 4 || os.Args[1] != "addr" || os.Args[2] != "request" {
		log.Fatal("invalid arguments")
	}

	config, err := apiconfig.LoadClientConfig("")
	if err != nil {
		log.Fatal(err)
	}

	client, err := clientv3.New(*config)
	if err != nil {
		log.Fatal(err)
	}

	IP, err := requestAddress(client, os.Args[3])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(IP)
}
func requestAddress(client clientv3.Interface, poolName string) (IP string, err error) {
	hostname, err := osutils.GetHostname()
	if err != nil {
		return "", err
	}

	var poolV4 []caliconet.IPNet

	poolsClient := client.IPPools()
	ipPool, err := poolsClient.Get(context.Background(), poolName, options.GetOptions{})
	if err != nil {
		return "", err
	}

	_, ipNet, err := caliconet.ParseCIDR(ipPool.Spec.CIDR)
	if err != nil {
		return "", err
	}

	poolV4 = []caliconet.IPNet{caliconet.IPNet{IPNet: ipNet.IPNet}}

	IPs, _, err := client.IPAM().AutoAssign(
		context.Background(),
		calicoipam.AutoAssignArgs{
			Num4:      1,
			Hostname:  hostname,
			IPv4Pools: poolV4,
		},
	)
	if err != nil {
		return "", err
	}
	if len(IPs) != 1 {
		return "", fmt.Errorf("unexpected number of IPs")
	}
	return string(IPs[0].IP), nil
}

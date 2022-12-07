package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/umegbewe/sshprobe/helpers"

	"fmt"
)

var instance []string

func GetInstances() ([]ec2.Instance, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := ec2.New(sess)

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running")},
			},
		},
	}

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return nil, fmt.Errorf("Couldn't list instances: %v", err)
	}

	var instances []ec2.Instance

	for _, res := range resp.Reservations {
		if res.Instances == nil {
			continue
		}

		for _, inst := range res.Instances {
			if inst == nil {
				continue
			}

			instance := ec2.Instance{
				PrivateIpAddress: inst.PrivateIpAddress,
				PublicIpAddress:  inst.PublicIpAddress,
				State:            inst.State,
				KeyName:          inst.KeyName,
			}

			instances = append(instances, instance)
		}
	}

	return instances, nil
}

func ssh(keyname string, user string, address string) {

	fmt.Println("ssh", "-o ConnectTimeout=10", user+"@"+address, "-i", "~/.ssh/"+keyname)
	cmd := exec.Command("ssh", "-o ConnectTimeout=10", user+"@"+address, "-i", "~/.ssh/"+keyname)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
	}
}

func Filter() []ec2.Instance {
	instances, err := GetInstances()
	if err != nil {
		fmt.Printf("Couldn't list instances: %v", err)
	}

	var instanceOutput strings.Builder
	for _, instance := range instances {
		instanceOutput.WriteString(fmt.Sprintf("%s | %s | %s | %s\n",
			helpers.StrOrDefault(instance.PrivateIpAddress, "None"),
			helpers.StrOrDefault(instance.PublicIpAddress, "None"),
			*instance.State.Name,
			helpers.StrOrDefault(instance.KeyName, "None"),
		))
	}

	// Convert the instances output to an io.Reader
	instancesReader := strings.NewReader(instanceOutput.String())

	var buf bytes.Buffer
	cmd := exec.Command("fzf", "--multi")
	cmd.Stdin = instancesReader
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		fmt.Printf("Couldn't call command: %v\n", err)
	}

	fzfOutput := buf.String()

	selectedInstances := strings.Split(fzfOutput, " | ")

	var filteredInstances []ec2.Instance
	for _, instance := range selectedInstances {
		privateIPAddress := strings.Split(instance, "|")[0]

		privateIPAddress = strings.TrimSpace(privateIPAddress)

		for _, i := range instances {
			if *i.PrivateIpAddress == privateIPAddress {
				filteredInstances = append(filteredInstances,  i)
			}
		}
	}

	return filteredInstances
}

func main() {
	selectedInstances := Filter()
	for _, instance := range selectedInstances {
		ssh(*instance.KeyName, "ubuntu", *instance.PrivateIpAddress)
	}
}

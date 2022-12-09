package main

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/umegbewe/sshprobe/helpers"

	"fmt"
)

type Instance struct {
	Name             string
	PublicIpAddress  *string
	PrivateIpAddress *string
	State            *ec2.InstanceState
	KeyName			 *string
}

var (
	instance  []string
	err       error
	user      = flag.String("user", "ubuntu", "Username to use")
	directory = flag.String("directory", "~/.ssh/", "Directory to find ssh keys")
)

func GetInstances() ([]*Instance, error) {
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

	var instances []*Instance

	for _, res := range resp.Reservations {
		if res.Instances == nil {
			continue
		}

		for _, inst := range res.Instances {
			if inst == nil {
				continue
			}

			instance := &Instance{
				Name:             helpers.GetTagName(inst),
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

func ssh(keyname string, user string, address string) error {
	var err error

	filename := *directory + "/" + keyname + ".pem"

	/* handle key pair's that might not have the .pem prefix*/
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		filename = *directory + "/" + keyname
	}

	fmt.Println("ssh", "-o ConnectTimeout=5", user+"@"+address, "-i", filename)
	cmd := exec.Command("ssh", "-o ConnectTimeout=5", user+"@"+address, "-i", filename)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		fmt.Println(err.Error())
	}

	return err
}

func Filter() []*Instance {
	instances, err := GetInstances()
	if err != nil {
		fmt.Printf("Couldn't list instances: %v", err)
	}

	var instanceOutput strings.Builder
	for _, instance := range instances {
		instanceOutput.WriteString(fmt.Sprintf("%s | %s | %s | %s | %s \n",
			helpers.StrOrDefault(instance.PrivateIpAddress, "None"),
			helpers.StrOrDefault(instance.PublicIpAddress, "None"),
			*instance.State.Name,
			helpers.StrOrDefault(instance.KeyName, "None"),
			instance.Name,
		))
	}

	// read buffer
	instancesReader := strings.NewReader(instanceOutput.String())

	var buf bytes.Buffer
	cmd := exec.Command("fzf", "--multi")
	cmd.Stdin = instancesReader
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf

	if err := cmd.Run(); cmd.ProcessState.ExitCode() == 130 { 
		} else if err != nil {
			fmt.Printf("Couldn't call command: %v\n", err)
	}
	
	fzfOutput := buf.String()

	selectedInstances := strings.Split(fzfOutput, " | ")

	var filteredInstances []*Instance
	for _, instance := range selectedInstances {
		privateIPAddress := strings.Split(instance, " | ")[0]

		privateIPAddress = strings.TrimSpace(privateIPAddress)

		for _, i := range instances {
			if *i.PrivateIpAddress == privateIPAddress {
				filteredInstances = append(filteredInstances, i)
			}
		}
	}

	return filteredInstances
}

func main() {
	flag.Parse()
	selectedInstances := Filter()
	for _, instance := range selectedInstances {
		err := ssh(*instance.KeyName, *user, *instance.PublicIpAddress)
		if err != nil {
			err = ssh(*instance.KeyName, *user, *instance.PrivateIpAddress)
		}
	}
}

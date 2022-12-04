package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"fmt"
	"log"
)

var instance []string

func ssh(keyname string, user string, address string) {

	fmt.Println("ssh", "-tt", user+"@"+address, "-i", "~/.ssh/"+keyname)
	cmd := exec.Command("ssh", "-tt", user+"@"+address, "-i", "~/.ssh/"+keyname)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
	}
}

// helper function
func StrOrDefault(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	} else {
		return *s
	}
}

func GetInstances() {

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
		fmt.Println("there was an error listing instances in", err.Error())
		log.Fatal(err.Error())
	}

	for _, res := range resp.Reservations {
		if res.Instances == nil {
			continue
		}

		for _, inst := range res.Instances {
			if inst == nil {
				fmt.Println("None")
				continue
			}

			instance = []string{
				*inst.PrivateIpAddress,
				StrOrDefault(inst.PublicIpAddress, "None"),
				*inst.State.Name,
				StrOrDefault(inst.KeyName, "None"),
			}

			fmt.Println(strings.Join(instance, " | "))
		}
	}
}

func Filter() string {

	stdout := os.Stdout
	fmt.Println("starting")
	r, w, _ := os.Pipe()
	os.Stdout = w

	GetInstances()

	ChnOut := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		ChnOut <- buf.String()
	}()

	w.Close()
	os.Stdout = stdout
	str := buf.String()

	cmd := exec.Command("fzf", "--multi")

	out := bytes.NewBufferString(str)

	cmd.Stdin = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); cmd.ProcessState.ExitCode() == 130 {
		return ""
	} else if err != nil {
		fmt.Errorf("Couldn't call fzf", err)
	}

	result := out.String()

	fmt.Println(result)

	return result

}

func main() {
	str := strings.Split(Filter(), " , ")

	fmt.Println(str)
}


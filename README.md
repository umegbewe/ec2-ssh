# **ec2-ssh**

Command-line tool to easily connect to Amazon EC2 instances via SSH. It uses **[fzf](https://github.com/junegunn/fzf)** to select instances to connect to, and it can be configured with various flags such as the SSH user, Region to find instances and directory where SSH keys are stored.

# Prerequisites
- Install [FZF](https://github.com/junegunn/fzf#installation)
- Setup AWS credentials https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/setup-credentials.html

## **Install**

Download [binary](https://github.com/umegbewe/ec2-ssh/releases)

## **Usage**

To use this tool, simply run the **`ec2-ssh`** command and follow the prompts to select the EC2 instance you want to connect to. By default, the tool will use the **`ubuntu`** user to connect, and it will look for SSH keys in the **`~/.ssh`** directory. You can use the following options to customize the tool's behavior:

```
-user: SSH user to login with. Default user is "ubuntu".
```
```
-directory: The directory where SSH keys are stored. Default is "~/.ssh".
```
```
-region: The region where the EC2 instances are located. Default is "us-east-1".
```

For example, to connect to an EC2 instance using the **`ec2-user`** user, SSH keys stored in the **`/home/ubuntu/keys`** directory and region **`us-west-2`**, you could use the following command:

```
ec2-ssh -user ec2-user -directory /home/ubuntu/keys -region us-west-2
```

## **License**

This tool is released under the MIT License. See **[LICENSE](https://github.com/umegbewe/ec2-ssh/blob/main/LICENSE)** for more information.
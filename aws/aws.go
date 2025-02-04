package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type AWS struct {
	Region         string `json:"region"`
	PemKeyFileName string `json:"pemkeyfilename"`
	AmiID          string `json:"amiid"`
	InstanceType   string `json:"instancetype"`
	Key            string `json:"key"`
	Secret         string `json:"secret"`
}
type EC2InstanceIP struct {
	InstanceID string
	PublicIP   string
	PrivateIP  string
}

func (a *AWS) createEc2Client() (*ec2.Client, error) {
	ctx := context.Background()
	customCreds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(a.Key, a.Secret, ""),
	)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(customCreds), config.WithRegion(a.Region))
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	return client, nil
}

func (a *AWS) getActiveEC2s(client *ec2.Client) (int, error) {
	ctx := context.Background()

	// Describe instances with the AUTO-BOX tag
	resp, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:AUTO-BOX"),
				Values: []string{"true"},
			},
			{
				Name:   aws.String("instance-state-code"),
				Values: []string{"16"},
			},
		},
	})

	if err != nil {
		return 0, err
	}

	return len(resp.Reservations), nil
}

func (a *AWS) createPEMFile(client *ec2.Client) error {
	ctx := context.Background()
	// Check if the key pair already exists
	existingKEY, err := client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("key-name"),
				Values: []string{a.PemKeyFileName},
			},
		},
	})
	if err != nil {
		return err
	}
	// If a security group with the given name exists, return its ID
	if len(existingKEY.KeyPairs) > 0 {
		return nil
	}

	resp, err := client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
		KeyName: aws.String(a.PemKeyFileName),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeKeyPair,
				Tags: []types.Tag{
					{
						Key:   aws.String("AUTO-BOX"),
						Value: aws.String("true"),
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	scriptsFolder := fmt.Sprintf("./%s", a.Region)
	// Ensure the directory exists
	err2 := os.MkdirAll(scriptsFolder, 0755)
	if err2 != nil {
		return err2
	}

	// fileName := fmt.Sprintf("%s.pem", keyName)
	fileName, err := filepath.Abs(filepath.Join(scriptsFolder, fmt.Sprintf("%s.pem", a.PemKeyFileName)))
	if err != nil {
		return err
	}
	err = os.WriteFile(fileName, []byte(*resp.KeyMaterial), 0400)
	if err != nil {
		return err
	}

	a.restrictWindowsFilePermissions(fileName)
	return nil
}

// will install in default VPC
func (a *AWS) createSecurityGroup(sgName, description string, client *ec2.Client) (string, error) {
	ctx := context.Background()

	existingGroups, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []string{sgName},
			},
		},
	})
	if err != nil {
		return "", err
	}

	// If a security group with the given name exists, return its ID
	if len(existingGroups.SecurityGroups) > 0 {
		return *existingGroups.SecurityGroups[0].GroupId, nil
	}

	resp, err := client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(sgName),
		Description: aws.String(description),
		// VpcId:       aws.String(vpcID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSecurityGroup,
				Tags: []types.Tag{
					{
						Key:   aws.String("AUTO-BOX"),
						Value: aws.String("true"),
					},
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	securityGroupID := *resp.GroupId

	rules := []struct {
		Protocol string
		Port     int32
	}{
		{"tcp", 80},   // HTTP
		{"tcp", 443},  // HTTPS
		{"tcp", 5901}, // VNC
		{"tcp", 22},   //telnet
	}

	for _, rule := range rules {
		_, err := client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(securityGroupID),
			IpPermissions: []types.IpPermission{
				{
					IpProtocol: aws.String(rule.Protocol),
					FromPort:   aws.Int32(rule.Port),
					ToPort:     aws.Int32(rule.Port),
					IpRanges: []types.IpRange{
						{CidrIp: aws.String("0.0.0.0/0")},
					},
				},
			},
		})
		if err != nil {
			return "", err
		}
	}

	// fmt.Printf("Security group created: %s\n", securityGroupID)
	return securityGroupID, nil
}

func (a *AWS) createEC2Instance(securityGroupID string, client *ec2.Client, batchT string) error {
	ctx := context.Background()

	resp, err := client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String(a.AmiID),
		InstanceType: types.InstanceType(a.InstanceType),
		KeyName:      aws.String(a.PemKeyFileName),
		SecurityGroupIds: []string{
			securityGroupID,
		},
		MinCount: aws.Int32(1),
		MaxCount: aws.Int32(1),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{
						Key:   aws.String("AUTO-BOX"),
						Value: aws.String("true"),
					},
					{
						Key:   aws.String("BatchTag"),
						Value: aws.String(batchT),
					},
				},
			},
		},
		InstanceMarketOptions: &types.InstanceMarketOptionsRequest{
			MarketType: types.MarketTypeSpot,
		},
	})
	if err != nil {
		return err
	}

	instanceID := *resp.Instances[0].InstanceId
	fmt.Printf("EC2 instance created: %s\n", instanceID)
	return nil
}

func (a *AWS) deleteEC2Instances(client *ec2.Client, batchT string) error {
	ctx := context.Background()

	// Describe instances with the AUTO-BOX tag
	resp, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:AUTO-BOX"),
				Values: []string{"true"},
			},
		},
	})
	if batchT != "" {
		resp, err = client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("tag:AUTO-BOX"),
					Values: []string{"true"},
				},
				{
					Name:   aws.String("tag:BatchTag"),
					Values: []string{batchT},
				},
			},
		})
	}
	if err != nil {
		return err
	}

	var instanceIDs []string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instanceIDs = append(instanceIDs, *instance.InstanceId)
		}
	}

	if len(instanceIDs) == 0 {
		fmt.Println("No EC2 instances with tag AUTO-BOX found.")
		return nil
	}

	// Terminate the instances
	_, err = client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: instanceIDs,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Terminated All EC2 instances: %v\n", instanceIDs)
	return nil
}

func (a *AWS) deletePEMFile(client *ec2.Client) error {
	ctx := context.Background()

	// Delete the key pair
	_, err := client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: aws.String(a.PemKeyFileName),
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *AWS) compileIPaddressesAws(client *ec2.Client, batchT string) (ips []string, fullEC2 []EC2InstanceIP, err error) {
	ctx := context.Background()

	// Describe EC2 instances
	resp, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:AUTO-BOX"),
				Values: []string{"true"},
			},
			{
				Name:   aws.String("tag:BatchTag"),
				Values: []string{batchT},
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	if batchT == "" {
		resp, err = client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("tag:AUTO-BOX"),
					Values: []string{"true"},
				},
			},
		})
		if err != nil {
			return nil, nil, err
		}
	}

	// Iterate over reservations and instances to collect IP addresses
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			ipInfo := EC2InstanceIP{
				InstanceID: *instance.InstanceId,
			}

			// Check if Public IP exists
			if instance.PublicIpAddress != nil {
				ipInfo.PublicIP = *instance.PublicIpAddress
				ips = append(ips, *instance.PublicIpAddress)
			}

			// Check if Private IP exists
			if instance.PrivateIpAddress != nil {
				ipInfo.PrivateIP = *instance.PrivateIpAddress
			}

			fullEC2 = append(fullEC2, ipInfo)
		}
	}

	return ips, fullEC2, nil
}

func (a *AWS) restrictWindowsFilePermissions(fileName string) error {
	// Get file attributes
	pointer, err := syscall.UTF16PtrFromString(fileName)
	if err != nil {
		return err
	}

	// Set file as readable only by the owner
	err = syscall.SetFileAttributes(pointer, syscall.FILE_ATTRIBUTE_READONLY)
	if err != nil {
		return err
	}

	return nil
}

func (a *AWS) cloneLambda() error {

	return nil
}

func (a *AWS) upgradeLambda() error {

	return nil
}

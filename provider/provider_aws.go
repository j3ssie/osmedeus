package provider

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
	"golang.org/x/crypto/ssh"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
)

// DefaultAWS set some default data for AWS provider
func (p *Provider) DefaultAWS() {
	p.Region = "ap-southeast-1"
	p.Size = "t2.medium"
	p.SecurityGroupName = "osmp-allow-root-access"
	p.SSHUser = p.ProviderConfig.Username

	if p.ProviderConfig.Username != "" {
		p.SSHUser = "admin"
	}
	if p.ProviderConfig.Region != "" {
		p.Region = p.ProviderConfig.Region
	}
	if p.ProviderConfig.Size != "" {
		p.Size = p.ProviderConfig.Size
	}
}

func (p *Provider) InitSessionAWS() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(p.Region),
		Credentials: credentials.NewStaticCredentials(p.AccessKeyId, p.SecretKey, ""),
	})

	return sess, err
}

func (p *Provider) ClientAWS() {
	client, err := p.InitSessionAWS()
	if err != nil {
		panic(err)
	}
	p.ProviderName = "aws"
	p.Client = client
}

func (p *Provider) ConvertClientAWS() *session.Session {
	sess, ok := p.Client.(*session.Session)
	if !ok {
		utils.ErrorF("error converting aws session %v", ok)
	}
	sess.Config.Region = aws.String(p.Region)
	return sess
}

func (p *Provider) AccountAWS() error {
	ceSvc := costexplorer.New(p.ConvertClientAWS())
	// Set the parameters for the query
	now := time.Now()
	start := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	params := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(start.Format("2006-01-02")),
			End:   aws.String(end.Format("2006-01-02")),
		},
		Granularity: aws.String(costexplorer.GranularityMonthly),
		Metrics:     []*string{aws.String("UnblendedCost")},
	}

	// Send the query and get the results
	// result, err := ceSvc.GetCostAndUsage(params)
	// if err != nil {
	// 	utils.ErrorF("Error getting cost and usage:", err)
	// 	return err
	// }

	// Send the query and get the results
	result, err := ceSvc.GetCostAndUsage(params)
	if err != nil {
		utils.ErrorF("Error getting cost and usage: %v", err)
		return err
	}

	// Print the total cost for the previous month
	cost := *result.ResultsByTime[0].Total["UnblendedCost"].Amount

	if !p.IsBackgroundCheck {
		utils.InforF("The total cost of AWS services for the this month was %s", color.HiRedString("$"+cost))
	}
	return nil
}

func (p *Provider) GetSSHKeyAWS() error {
	ec2Svc := ec2.New(p.ConvertClientAWS())
	// Get a list of all the key pairs in the account
	keyPairsOutput, err := ec2Svc.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		utils.ErrorF("Error describing key pairs: %v", err)
		return err
	}

	pubKeyBytes := []byte(p.SSHPublicKey)
	// Parse the key, other info ignored
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyBytes)
	if err != nil {
		utils.ErrorF("%v", err)
		return err
	}
	hash := sha256.Sum256(pubKey.Marshal())
	sshHash := base64.StdEncoding.EncodeToString(hash[:])

	// Check if your SSH key is present in the list
	for _, keyPair := range keyPairsOutput.KeyPairs {
		// found the same key name but different key fingerprint
		if *keyPair.KeyName == p.SSHKeyName && *keyPair.KeyFingerprint != sshHash {
			// Delete the key pair
			_, err := ec2Svc.DeleteKeyPair(&ec2.DeleteKeyPairInput{
				KeyName: aws.String(*keyPair.KeyName),
			})
			if err != nil {
				utils.ErrorF("%v", err)
			}
			utils.InforF("Successfully deleted key pair %s", color.HiBlueString(*keyPair.KeyName))
		}

		if *keyPair.KeyFingerprint == sshHash {
			p.SSHKeyID = cast.ToString(*keyPair.KeyPairId)
			p.SSHKeyFound = true
			break
		}

	}

	if p.SSHKeyFound {
		utils.DebugF("Your SSH key was found in the account: %v -- %v", color.HiCyanString(p.SSHKeyName), color.HiCyanString(p.SSHKeyID))
		return nil
	}

	// Import the SSH key into your AWS account
	result, err := ec2Svc.ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           aws.String(p.SSHKeyName),
		PublicKeyMaterial: []byte(p.SSHPublicKey),
	})
	if err != nil {
		utils.ErrorF("Error create key pairs: %v", err)
		return err
	}
	utils.DebugF("Successfully imported SSH key: %v -- %v", color.HiCyanString(*result.KeyName), color.HiCyanString(*result.KeyPairId))
	p.SSHKeyID = cast.ToString(*result.KeyPairId)
	p.SSHKeyFound = true

	return nil
}

func (p *Provider) ListSnapshotAWS() error {
	svc := ec2.New(p.ConvertClientAWS())

	// listing only image that own by you
	self := "self"
	ownImages := &ec2.DescribeImagesInput{Owners: []*string{&self}}
	result, err := svc.DescribeImages(ownImages)
	if err != nil {
		utils.ErrorF("err: Unable to list images, %v", err)
		return err
	}

	for _, item := range result.Images {
		if strings.HasPrefix(*item.Name, libs.SNAPSHOT) {
			p.OldSnapShotID = append(p.OldSnapShotID, *item.ImageId)
		}

		if strings.TrimSpace(*item.Name) == strings.TrimSpace(p.SnapshotName) {
			utils.DebugF("Found base image snapshot with ID: %s", color.HiBlueString(*item.ImageId))
			p.SnapshotID = *item.ImageId
			p.SnapshotName = *item.Name
			p.SnapshotFound = true
		}
	}

	return nil
}

func (p *Provider) DeleteImageAWS(id string) error {
	if p.SnapshotID == "" {
		return nil
	}
	svc := ec2.New(p.ConvertClientAWS())
	deletedImage := &ec2.DeregisterImageInput{ImageId: &p.SnapshotID}
	_, err := svc.DeregisterImage(deletedImage)
	if err != nil {
		utils.ErrorF("err: Unable to delete snapshot: %v -- %v", id, err)
		return err
	}
	utils.InforF("Deleted image ID: %v", color.HiRedString(p.SnapshotID))
	p.DeleteSnapshotAWS()
	return nil
}

func (p *Provider) DeleteSnapshotAWS() error {
	svc := ec2.New(p.ConvertClientAWS())

	// List all snapshots
	result, err := svc.DescribeSnapshots(&ec2.DescribeSnapshotsInput{
		OwnerIds:   []*string{aws.String("self")},
		MaxResults: aws.Int64(1000),
	})
	if err != nil {
		utils.ErrorF("Error listing snapshots: %v", err)
		return err
	}

	var snapshotIDs []string
	for _, snapshot := range result.Snapshots {

		for _, tag := range snapshot.Tags {
			if *tag.Key == "Name" && *tag.Value == "Osmedeus Premium Image" {
				snapshotIDs = append(snapshotIDs, *snapshot.SnapshotId)
			}
		}
	}

	for _, snapshotID := range snapshotIDs {
		_, err = svc.DeleteSnapshot(&ec2.DeleteSnapshotInput{
			SnapshotId: aws.String(snapshotID),
		})
		if err != nil {
			utils.ErrorF("Error deleting snapshot: %v -- %v", snapshotID, err)
		} else {
			utils.DebugF("Delted snapshot ID: %v", color.HiRedString(snapshotID))
		}
	}

	return nil
}

func (p *Provider) ListInstanceAWS() error {
	svc := ec2.New(p.ConvertClientAWS())
	result, err := svc.DescribeInstances(nil)
	if err != nil {
		utils.ErrorF("err: Unable to list ec2 instances: %v", err)
		return err
	}

	var numberOfInstance int
	for i := range result.Reservations {
		for _, instance := range result.Reservations[i].Instances {
			if *instance.State.Name != "running" {
				continue
			}
			numberOfInstance += 1

			launchTime := *instance.LaunchTime
			creationDate := launchTime.Format(time.RFC1123)
			parsedInstance := Instance{
				InstanceID:   cast.ToString(*instance.InstanceId),
				IPAddress:    *instance.PublicIpAddress,
				InstanceName: *instance.State.Name,
				ImageID:      cast.ToString(*instance.ImageId),
				// ImageName: instance.Image.Name,
				// Region: instance.Region.Slug,
				// Region:    *instance.Architecture,
				// Memory:       cast.ToString(instance.Memory),
				// CPU:          cast.ToString(instance.Vcpus),
				// Disk:         cast.ToString(instance.Disk),
				Status:       *instance.State.Name,
				CreatedAt:    cast.ToString(creationDate),
				InputName:    "",
				ProviderName: "aws",
			}

			p.Instances = append(p.Instances, parsedInstance)
		}
	}

	utils.InforF("Found %v running instances", color.HiMagentaString("%v", numberOfInstance))
	// check if we reach max instance number
	if p.InstanceLimit > 0 {
		if len(p.Instances) >= p.InstanceLimit {
			p.Available = false
		}
	}

	return nil
}

func (p *Provider) CreateInstanceAWS(InstanceName string) (instanctID string, err error) {
	svc := ec2.New(p.ConvertClientAWS())
	p.CreateSecurityGroup()

	// Set the parameters for the instance
	params := &ec2.RunInstancesInput{
		// InstanceName: aws.String(name),
		ImageId:      aws.String(p.SnapshotID), // Replace with the ID of the image you want to use
		InstanceType: aws.String(p.Size),       // Specify the instance type like t2.micro
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      aws.String(p.SSHKeyName),
		SecurityGroups: []*string{
			aws.String(p.SecurityGroupName),
		},
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(InstanceName),
					},
				},
			},
		},
	}

	// Create the instance
	result, err := svc.RunInstances(params)
	if err != nil {
		utils.ErrorF("Error creating instance: %v", err)
		return
	}

	// Get the instance ID
	instanctID = *result.Instances[0].InstanceId
	utils.InforF("Successfully Created Instance ID: %v -- %v", color.HiBlueString(instanctID), color.HiBlueString(InstanceName))
	utils.DebugF("Waiting for the instance %v to be ready...", color.HiBlueString(instanctID))

	time.Sleep(60 * time.Second)
	// Get the instance state
	for i := 0; i < 10; i++ {
		if p.InstanceReady(instanctID) == nil {
			return instanctID, nil
		}
		time.Sleep(60 * time.Second)
	}

	return instanctID, nil
}

func (p *Provider) InstanceReady(instanceID string) error {
	svc := ec2.New(p.ConvertClientAWS())

	// Describe the instance
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}
	_, err := svc.DescribeInstances(params)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) InstanceReboot(instanceID string) error {
	svc := ec2.New(p.ConvertClientAWS())

	// Create the input for the RebootInstances operation
	params := &ec2.RebootInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}

	// Call the RebootInstances operation
	_, err := svc.RebootInstances(params)
	if err != nil {
		fmt.Println("Error rebooting instance:", err)
		return err
	}
	return nil
}

func (p *Provider) DeleteInstanceAWS(id string) error {
	svc := ec2.New(p.ConvertClientAWS())

	// Set the parameters for the instance
	params := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(id), // Replace with the ID of the instance you want to delete
		},
	}

	// Delete the instance
	_, err := svc.TerminateInstances(params)
	if err != nil {
		utils.ErrorF("Error deleting instance: %v", err)
		return err
	}

	utils.InforF("Successfully Deleted instance ID: %v", color.HiRedString(id))
	return nil
}

func (p *Provider) InstanceInfoAWS(id string) (Instance, error) {
	var parsedInstance Instance

	svc := ec2.New(p.ConvertClientAWS())

	// Set the parameters for the instance
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(id),
		},
	}

	// Get the instance information
	result, err := svc.DescribeInstances(params)
	if err != nil {
		utils.ErrorF("Error getting instance information: %v", err)
		return parsedInstance, err
	}

	// Print the instance information
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

			launchTime := *instance.LaunchTime
			creationDate := launchTime.Format(time.RFC1123)

			var instanceName string
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					instanceName = *tag.Value
					break
				}
			}

			parsedInstance = Instance{
				InstanceID:   cast.ToString(*instance.InstanceId),
				IPAddress:    *instance.PublicIpAddress,
				InstanceName: instanceName,
				ImageID:      cast.ToString(*instance.ImageId),
				Status:       *instance.State.Name,
				CreatedAt:    cast.ToString(creationDate),
				InputName:    "",
				ProviderName: "aws",
			}

		}
	}

	p.CreatedInstance = parsedInstance
	utils.DebugF("Instance ID Info: %v -- %v -- %v", color.HiBlueString(p.CreatedInstance.InstanceID), p.CreatedInstance.InstanceName, p.CreatedInstance.IPAddress)
	return parsedInstance, nil
}

func (p *Provider) CreateSecurityGroup() error {
	svc := ec2.New(p.ConvertClientAWS())

	// Set the parameters for the security group
	params := &ec2.DescribeSecurityGroupsInput{
		GroupNames: []*string{
			aws.String(p.SecurityGroupName), // Replace with the ID of the security group you want to check
		},
	}

	// Get the security group information
	scGroups, err := svc.DescribeSecurityGroups(params)
	if err == nil {
		// Print the security group information
		for _, group := range scGroups.SecurityGroups {
			if *group.GroupName == p.SecurityGroupName {
				p.SecurityGroupID = *group.GroupId
				utils.DebugF("Security Group allow root access has been found: %v", color.HiBlueString(p.SecurityGroupID))
				return nil
			}
		}
	}

	// only create if not found

	// Create the security group
	result, err := svc.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String("osmp-allow-root-access"),
		Description: aws.String("Security group for allowing root access to EC2 instances"),
	})
	if err != nil {
		utils.ErrorF("Error creating security group: %v", err)
		return err
	}

	// Add a rule to the security group to allow SSH access from any IP address
	_, err = svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(*result.GroupId),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(22),
				ToPort:     aws.Int64(22),
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{CidrIp: aws.String("0.0.0.0/0")},
				},
			},
		},
	})
	if err != nil {
		utils.ErrorF("Error adding rule to security group: %v", err)
		return err
	}

	p.SecurityGroupID = *result.GroupId
	utils.DebugF("Security Group allow root access has been found: %v", color.HiBlueString(p.SecurityGroupID))
	return nil
}

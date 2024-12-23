package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type Client struct {
	EC2 *ec2.Client
}

func newClient() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return &Client{
		EC2: ec2.NewFromConfig(cfg),
	}, nil
}

type Instance struct {
	Name  string
	ID    string
	Type  string
	State string
}

func ListInstances() ([]Instance, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	resp, err := c.EC2.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	var instances []Instance
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			name := ""
			for _, tag := range inst.Tags {
				if *tag.Key == "Name" {
					name = *tag.Value
					break
				}
			}

			instances = append(instances, Instance{
				ID:    *inst.InstanceId,
				Type:  string(inst.InstanceType),
				State: string(inst.State.Name),
				Name:  name,
			})
		}
	}
	return instances, nil
}

package internal

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"sort"
)

type Instance struct {
	ID        string `json:"instanceId"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	State     string `json:"state"`
	Platform  string `json:"platform"`
	Supported bool
}

var (
	LIST_INSTANCES     = []string{"ec2", "describe-instances", "--query", "Reservations[].Instances[].{instanceId:InstanceId,name:Tags[?Key==`Name`].Value|[0],type:InstanceType,state:State.Name,platform:Platform}"}
	LIST_SSM_INSTANCES = []string{"ssm", "describe-instance-information", "--query", "InstanceInformationList[].InstanceId"}
)

func ListInstances(profile string) ([]Instance, error) {

	cmd := exec.Command("aws", LIST_INSTANCES...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var ec2Instances []Instance
	if err := json.Unmarshal(output, &ec2Instances); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w\noutput: %s", err, string(output))
	}

	cmd = exec.Command("aws", LIST_SSM_INSTANCES...)
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var ssmInstances []string
	if err := json.Unmarshal(output, &ssmInstances); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w\noutput: %s", err, string(output))
	}

	var tableInstances []Instance
	for _, inst := range ec2Instances {

		if slices.Contains(ssmInstances, inst.ID) {
			inst.Supported = true
		} else {
			inst.Supported = false
		}
		if inst.Platform == "" {
			inst.Platform = "Linux"
		}
		tableInstances = append(tableInstances, inst)
	}

	sort.Slice(tableInstances, func(i, j int) bool {

		if tableInstances[i].Supported && !tableInstances[j].Supported {
			return true
		} else if !tableInstances[i].Supported && tableInstances[j].Supported {
			return false
		}
		return tableInstances[i].Name < tableInstances[j].Name
	})

	return tableInstances, nil
}

package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"sort"
	"strings"
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

	var outBuf, errBuf bytes.Buffer
	var output []byte

	cmd := exec.Command("aws", LIST_INSTANCES...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		return nil, fmt.Errorf("failed to execute command: %s", errMsg)
	} else {
		output = outBuf.Bytes()
	}
	outBuf.Reset()
	errBuf.Reset()

	var ec2Instances []Instance
	if err := json.Unmarshal(output, &ec2Instances); err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		return nil, errors.New(errMsg)
	}

	cmd = exec.Command("aws", LIST_SSM_INSTANCES...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		return nil, errors.New(errMsg)
	} else {
		output = outBuf.Bytes()
	}
	outBuf.Reset()
	errBuf.Reset()

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

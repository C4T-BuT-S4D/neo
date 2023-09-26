package models

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type Target struct {
	ID string
	IP string
}

func NewTarget(teamID, teamIP string) *Target {
	return &Target{
		ID: teamID,
		IP: teamIP,
	}
}

func (i *Target) String() string {
	return fmt.Sprintf("Target(id=%s, ip=%s)", i.ID, i.IP)
}

func (i *Target) MetricLabels() prometheus.Labels {
	return prometheus.Labels{"target_id": i.ID, "target_ip": i.IP}
}

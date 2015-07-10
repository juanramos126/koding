package tunnelproxymanager

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	healthy   = "Healthy"
	inService = "InService"
)

var (
	errAutoscalingNotSet = errors.New("autoscaling is not set")
	errASGNameNotSet     = errors.New("ASG name is not set")
)

// AttachNotificationToAutoScaling attaches topic to autoscaling group, multiple
// calls to this function will result in same
func (l *LifeCycle) AttachNotificationToAutoScaling() error {
	log := l.log.New("AutoScaling")

	if l.topicARN == nil {
		return errTopicARNNotSet
	}

	if l.autoscaling == nil {
		return errAutoscalingNotSet
	}

	if l.asgName == nil {
		return errASGNameNotSet
	}

	log.Debug("attaching SNS topic (%s) to AutoScaling (%s) for listening scaling events", *l.topicARN, *l.asgName)

	// PutNotificationConfiguration is idempotent
	if _, err := l.autoscaling.PutNotificationConfiguration(
		&autoscaling.PutNotificationConfigurationInput{
			AutoScalingGroupName: l.asgName,
			NotificationTypes: []*string{
				aws.String("autoscaling:EC2_INSTANCE_LAUNCH"),
				aws.String("autoscaling:EC2_INSTANCE_LAUNCH_ERROR"),
				aws.String("autoscaling:EC2_INSTANCE_TERMINATE"),
				aws.String("autoscaling:EC2_INSTANCE_TERMINATE_ERROR"),
			},
			TopicARN: l.topicARN,
		},
	); err != nil {
		return err
	}

	log.Debug("notification configuration is completed successfully")
	return nil
}

// GetAutoScalingOperatingIPs gets the Healthy and InService servers' IP
// addresses of an autscaling group
func (l *LifeCycle) GetAutoScalingOperatingIPs() ([]*string, error) {
	if l.asgName == nil {
		return nil, errASGNameNotSet
	}

	if l.autoscaling == nil {
		return nil, errAutoscalingNotSet
	}

	// get instances of our autoscaling group
	asResp, err := l.autoscaling.DescribeAutoScalingGroups(
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{l.asgName},
		},
	)
	if err != nil {
		return nil, err
	}

	if asResp == nil || asResp.AutoScalingGroups == nil {
		return nil, errors.New("describe asg response is malformed")
	}

	healthyInstances := filterHealthyInstances(asResp)
	if len(healthyInstances) == 0 {
		return nil, errors.New("no instances are healthy and in service")
	}

	insResp, err := l.ec2.DescribeInstances(
		&ec2.DescribeInstancesInput{
			InstanceIDs: healthyInstances,
		},
	)
	if err != nil {
		return nil, err
	}

	if insResp == nil || insResp.Reservations == nil {
		return nil, errors.New("describe instances response is malformed")
	}
	return mapPublicIps(insResp), nil
}

func filterHealthyInstances(asResp *autoscaling.DescribeAutoScalingGroupsOutput) []*string {
	healthyInstances := make([]*string, 0)
	for _, asg := range asResp.AutoScalingGroups {
		for _, instance := range asg.Instances {
			if *instance.HealthStatus == healthy &&
				*instance.LifecycleState == inService {
				healthyInstances = append(healthyInstances, instance.InstanceID)
			}
		}
	}
	return healthyInstances
}

func mapPublicIps(insResp *ec2.DescribeInstancesOutput) []*string {
	publicIps := make([]*string, 0)
	for _, reservation := range insResp.Reservations {
		for _, instance := range reservation.Instances {
			publicIps = append(publicIps, instance.PublicIPAddress)
		}
	}
	return publicIps
}
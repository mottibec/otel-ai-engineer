package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	"github.com/mottibechhofer/otel-ai-engineer/tools/grafana/deployers"
)

// DeployGrafanaInput represents the input for deploying Grafana
type DeployGrafanaInput struct {
	TargetType    string                 `json:"target_type"`
	InstanceName  string                 `json:"instance_name"`
	AdminUser     string                 `json:"admin_user"`
	AdminPassword string                 `json:"admin_password"`
	Parameters    map[string]interface{} `json:"parameters"`
}

// GetDeployGrafanaTool creates a tool for deploying Grafana
func GetDeployGrafanaTool() tools.Tool {
	return tools.Tool{
		Name:        "deploy_grafana",
		Description: "Deploys a new Grafana instance to the specified target (docker, kubernetes, or remote). The Grafana instance can be configured to connect to OpenTelemetry collectors automatically.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"target_type": map[string]interface{}{
					"type":        "string",
					"description": "Deployment target type: 'docker', 'kubernetes', or 'remote'",
					"enum":        []string{"docker", "kubernetes", "remote"},
				},
				"instance_name": map[string]interface{}{
					"type":        "string",
					"description": "Name for the Grafana instance",
				},
				"admin_user": map[string]interface{}{
					"type":        "string",
					"description": "Admin username (default: admin)",
				},
				"admin_password": map[string]interface{}{
					"type":        "string",
					"description": "Admin password (default: admin)",
				},
				"parameters": map[string]interface{}{
					"type":        "object",
					"description": "Target-specific deployment parameters. For docker: network (default: 'otel-network'), image (default: 'grafana/grafana:latest'), port (default: '3000').",
				},
			},
			Required: []string{"target_type", "instance_name"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input DeployGrafanaInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Get the appropriate deployer
			deployer, err := getDeployer(deployers.TargetType(input.TargetType))
			if err != nil {
				return nil, fmt.Errorf("unsupported target type %s: %w", input.TargetType, err)
			}

			// Deploy Grafana
			config := deployers.GrafanaDeploymentConfig{
				TargetType:    deployers.TargetType(input.TargetType),
				InstanceName: input.InstanceName,
				AdminUser:     input.AdminUser,
				AdminPassword: input.AdminPassword,
				Parameters:    input.Parameters,
			}

			result, err := deployer.Deploy(config)
			if err != nil {
				return nil, fmt.Errorf("deployment failed: %w", err)
			}

			return result, nil
		},
	}
}

// getDeployer returns the appropriate deployer for the target type
func getDeployer(targetType deployers.TargetType) (deployers.GrafanaDeployer, error) {
	switch targetType {
	case deployers.TargetDocker:
		return deployers.NewDockerDeployer()
	case deployers.TargetK8s:
		return deployers.NewKubernetesDeployer()
	case deployers.TargetRemote:
		return deployers.NewRemoteDeployer()
	default:
		return nil, fmt.Errorf("unknown target type: %s", targetType)
	}
}

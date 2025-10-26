package otel

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
	"github.com/mottibechhofer/otel-ai-engineer/tools/otel/deployers"
)

// DeployCollectorInput represents the input for deploying a collector
type DeployCollectorInput struct {
	TargetType     string                 `json:"target_type"`
	CollectorName  string                 `json:"collector_name"`
	YAMLConfig     string                 `json:"yaml_config"`
	Parameters     map[string]interface{} `json:"parameters"`
}

// GetDeployCollectorTool creates a tool for deploying a collector
func GetDeployCollectorTool() tools.Tool {
	return tools.Tool{
		Name:        "deploy_otel_collector",
		Description: "Deploys a new OpenTelemetry collector instance to the specified target (docker, remote, kubernetes, or local). The collector will automatically connect to the Lawrence OpAMP server. Currently supports docker deployment.",
		Schema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"target_type": map[string]interface{}{
					"type":        "string",
					"description": "Deployment target type: 'docker', 'remote', 'kubernetes', or 'local'",
					"enum":        []string{"docker", "remote", "kubernetes", "local"},
				},
				"collector_name": map[string]interface{}{
					"type":        "string",
					"description": "Name for the collector instance",
				},
				"yaml_config": map[string]interface{}{
					"type":        "string",
					"description": "YAML configuration for the collector. Must include OpAMP extension configuration pointing to Lawrence server.",
				},
				"parameters": map[string]interface{}{
					"type":        "object",
					"description": "Target-specific deployment parameters. For docker: network (default: 'otel-network'), image (default: 'otel/opentelemetry-collector-contrib:latest'), lawrence_url (default: 'http://lawrence:4320').",
				},
			},
			Required: []string{"target_type", "collector_name", "yaml_config"},
		},
		Handler: func(inputJSON json.RawMessage) (interface{}, error) {
			var input DeployCollectorInput
			if err := json.Unmarshal(inputJSON, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// Get the appropriate deployer
			deployer, err := getDeployer(deployers.TargetType(input.TargetType))
			if err != nil {
				return nil, fmt.Errorf("unsupported target type %s: %w", input.TargetType, err)
			}

			// Deploy the collector
			config := deployers.DeploymentConfig{
				TargetType:     deployers.TargetType(input.TargetType),
				CollectorName: input.CollectorName,
				YAMLConfig:     input.YAMLConfig,
				Parameters:     input.Parameters,
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
func getDeployer(targetType deployers.TargetType) (deployers.Deployer, error) {
	switch targetType {
	case deployers.TargetDocker:
		return deployers.NewDockerDeployer()
	case deployers.TargetRemote:
		return nil, fmt.Errorf("remote deployment not yet implemented")
	case deployers.TargetK8s:
		return nil, fmt.Errorf("kubernetes deployment not yet implemented")
	case deployers.TargetLocal:
		return nil, fmt.Errorf("local process deployment not yet implemented")
	default:
		return nil, fmt.Errorf("unknown target type: %s", targetType)
	}
}


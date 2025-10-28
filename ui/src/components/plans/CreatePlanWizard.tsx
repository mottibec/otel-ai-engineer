import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { apiClient } from "../../services/api";
import type { ObservabilityPlan } from "../../types/plan";

interface CreatePlanWizardProps {
  isOpen: boolean;
  onClose: () => void;
  onPlanCreated?: (plan: ObservabilityPlan) => void;
}

interface WizardStepData {
  name: string;
  description: string;
  environment: string;
  services: Array<{
    name: string;
    language: string;
    framework: string;
    path: string;
  }>;
  infrastructure: Array<{
    name: string;
    type: string;
    host: string;
    receiver: string;
  }>;
  pipelines: Array<{
    name: string;
    collectorId: string;
  }>;
  backends: Array<{
    name: string;
    type: string;
    url: string;
  }>;
}

const INITIAL_STEP_DATA: WizardStepData = {
  name: "",
  description: "",
  environment: "",
  services: [],
  infrastructure: [],
  pipelines: [],
  backends: [],
};

export function CreatePlanWizard({ isOpen, onClose, onPlanCreated }: CreatePlanWizardProps) {
  const [currentStep, setCurrentStep] = useState(1);
  const [data, setData] = useState<WizardStepData>(INITIAL_STEP_DATA);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const totalSteps = 6;

  const handleNext = () => {
    if (currentStep < totalSteps) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleBack = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleSubmit = async () => {
    try {
      setLoading(true);
      setError(null);

      const plan = await apiClient.createPlan({
        name: data.name,
        description: data.description,
        environment: data.environment,
        status: "draft",
      });

      if (onPlanCreated) {
        onPlanCreated(plan);
      }

      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create plan");
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setCurrentStep(1);
    setData(INITIAL_STEP_DATA);
    setError(null);
    onClose();
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return (
          <div className="space-y-4">
            <div>
              <Label htmlFor="name">Plan Name *</Label>
              <Input
                id="name"
                value={data.name}
                onChange={(e) => setData({ ...data, name: e.target.value })}
                placeholder="My Observability Plan"
                required
              />
            </div>
            <div>
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={data.description}
                onChange={(e) => setData({ ...data, description: e.target.value })}
                placeholder="Describe what this observability plan will cover"
                rows={4}
              />
            </div>
            <div>
              <Label htmlFor="environment">Environment</Label>
              <Input
                id="environment"
                value={data.environment}
                onChange={(e) => setData({ ...data, environment: e.target.value })}
                placeholder="production, staging, development"
              />
            </div>
          </div>
        );

      case 2:
        return (
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Add Service</CardTitle>
                <CardDescription>Add a service to instrument</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <Label htmlFor="service-name">Service Name</Label>
                  <Input
                    id="service-name"
                    placeholder="user-service"
                    onKeyPress={(e) => {
                      if (e.key === "Enter") {
                        const input = e.target as HTMLInputElement;
                        if (input.value) {
                          setData({
                            ...data,
                            services: [
                              ...data.services,
                              {
                                name: input.value,
                                language: "",
                                framework: "",
                                path: "",
                              },
                            ],
                          });
                          input.value = "";
                        }
                      }
                    }}
                  />
                </div>
                {data.services.map((service, idx) => (
                  <Card key={idx} className="p-4">
                    <div className="flex items-center justify-between">
                      <span className="font-medium">{service.name}</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          const newServices = [...data.services];
                          newServices.splice(idx, 1);
                          setData({ ...data, services: newServices });
                        }}
                      >
                        Remove
                      </Button>
                    </div>
                  </Card>
                ))}
              </CardContent>
            </Card>
            <p className="text-sm text-muted-foreground">
              Enter a service name and press Enter to add it to the list
            </p>
          </div>
        );

      case 3:
        return (
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Add Infrastructure Component</CardTitle>
                <CardDescription>Add infrastructure to monitor</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Component Name</Label>
                  <Input
                    placeholder="postgres-db"
                    onKeyPress={(e) => {
                      if (e.key === "Enter") {
                        const input = e.target as HTMLInputElement;
                        if (input.value) {
                          setData({
                            ...data,
                            infrastructure: [
                              ...data.infrastructure,
                              {
                                name: input.value,
                                type: "",
                                host: "",
                                receiver: "",
                              },
                            ],
                          });
                          input.value = "";
                        }
                      }
                    }}
                  />
                </div>
                {data.infrastructure.map((infra, idx) => (
                  <Card key={idx} className="p-4">
                    <div className="flex items-center justify-between">
                      <span className="font-medium">{infra.name}</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          const newInfra = [...data.infrastructure];
                          newInfra.splice(idx, 1);
                          setData({ ...data, infrastructure: newInfra });
                        }}
                      >
                        Remove
                      </Button>
                    </div>
                  </Card>
                ))}
              </CardContent>
            </Card>
          </div>
        );

      case 4:
        return (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Collector pipelines will be configured automatically based on your services and infrastructure.
            </p>
          </div>
        );

      case 5:
        return (
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Add Backend</CardTitle>
                <CardDescription>Add an observability backend</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Backend URL</Label>
                  <Input
                    placeholder="https://grafana.example.com"
                    onKeyPress={(e) => {
                      if (e.key === "Enter") {
                        const input = e.target as HTMLInputElement;
                        if (input.value) {
                          setData({
                            ...data,
                            backends: [
                              ...data.backends,
                              {
                                name: `Backend ${data.backends.length + 1}`,
                                type: "grafana",
                                url: input.value,
                              },
                            ],
                          });
                          input.value = "";
                        }
                      }
                    }}
                  />
                </div>
                {data.backends.map((backend, idx) => (
                  <Card key={idx} className="p-4">
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="font-medium">{backend.name}</span>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            const newBackends = [...data.backends];
                            newBackends.splice(idx, 1);
                            setData({ ...data, backends: newBackends });
                          }}
                        >
                          Remove
                        </Button>
                      </div>
                      <p className="text-sm text-muted-foreground">{backend.url}</p>
                    </div>
                  </Card>
                ))}
              </CardContent>
            </Card>
          </div>
        );

      case 6:
        return (
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Review Your Plan</CardTitle>
                <CardDescription>Confirm the details before creating</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <strong>Name:</strong> {data.name}
                </div>
                <div>
                  <strong>Description:</strong> {data.description || "None"}
                </div>
                <div>
                  <strong>Environment:</strong> {data.environment || "None"}
                </div>
                <div>
                  <strong>Services:</strong> {data.services.length}
                </div>
                <div>
                  <strong>Infrastructure Components:</strong> {data.infrastructure.length}
                </div>
                <div>
                  <strong>Backends:</strong> {data.backends.length}
                </div>
              </CardContent>
            </Card>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Observability Plan</DialogTitle>
          <DialogDescription>
            Step {currentStep} of {totalSteps}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Progress indicator */}
          <div className="flex items-center justify-between mb-4">
            {Array.from({ length: totalSteps }).map((_, idx) => (
              <div
                key={idx}
                className={`flex-1 h-2 mx-1 rounded ${
                  idx < currentStep ? "bg-primary" : "bg-muted"
                }`}
              />
            ))}
          </div>

          {/* Step content */}
          <div className="min-h-[300px]">{renderStepContent()}</div>

          {/* Error message */}
          {error && (
            <div className="p-4 bg-destructive/10 text-destructive rounded-md text-sm">
              {error}
            </div>
          )}

          {/* Navigation buttons */}
          <div className="flex items-center justify-between">
            <div>
              {currentStep > 1 && (
                <Button variant="outline" onClick={handleBack}>
                  Back
                </Button>
              )}
            </div>
            <div className="flex gap-2">
              {currentStep < totalSteps ? (
                <Button onClick={handleNext} disabled={!data.name}>
                  Next
                </Button>
              ) : (
                <Button onClick={handleSubmit} disabled={loading || !data.name}>
                  {loading ? "Creating..." : "Create Plan"}
                </Button>
              )}
              <Button variant="outline" onClick={handleClose}>
                Cancel
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}


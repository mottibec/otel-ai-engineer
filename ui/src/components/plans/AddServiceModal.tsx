import { useState } from "react";
import { apiClient } from "@/services/api";
import type { InstrumentedService } from "@/types/plan";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Loader2 } from "lucide-react";

interface AddServiceModalProps {
  isOpen: boolean;
  onClose: () => void;
  planId: string;
  onServiceAdded: () => void;
}

export function AddServiceModal({
  isOpen,
  onClose,
  planId,
  onServiceAdded,
}: AddServiceModalProps) {
  const [serviceName, setServiceName] = useState("");
  const [language, setLanguage] = useState("");
  const [framework, setFramework] = useState("");
  const [targetPath, setTargetPath] = useState("");
  const [gitRepoURL, setGitRepoURL] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!serviceName.trim()) {
      setError("Service name is required");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const service: Partial<InstrumentedService> = {
        plan_id: planId,
        service_name: serviceName.trim(),
        language: language || undefined,
        framework: framework || undefined,
        target_path: targetPath || undefined,
        git_repo_url: gitRepoURL || undefined,
        status: "pending",
        sdk_version: "",
        config_file: "",
        code_changes_summary: "",
        exporter_endpoint: "",
      };

      await apiClient.createService(planId, service as InstrumentedService);
      onServiceAdded();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create service");
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!submitting) {
      setServiceName("");
      setLanguage("");
      setFramework("");
      setTargetPath("");
      setGitRepoURL("");
      setError(null);
      onClose();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Add Service to Plan</DialogTitle>
          <DialogDescription>
            Add a new service that needs instrumentation. You can connect a git repository if code changes are needed.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Service Name */}
          <div className="space-y-2">
            <Label htmlFor="serviceName">Service Name *</Label>
            <Input
              id="serviceName"
              value={serviceName}
              onChange={(e) => setServiceName(e.target.value)}
              placeholder="e.g., user-service, api-service"
              disabled={submitting}
              required
            />
          </div>

          {/* Language */}
          <div className="space-y-2">
            <Label htmlFor="language">Language</Label>
            <Select value={language} onValueChange={setLanguage} disabled={submitting}>
              <SelectTrigger id="language">
                <SelectValue placeholder="Select language" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="go">Go</SelectItem>
                <SelectItem value="python">Python</SelectItem>
                <SelectItem value="javascript">JavaScript</SelectItem>
                <SelectItem value="typescript">TypeScript</SelectItem>
                <SelectItem value="java">Java</SelectItem>
                <SelectItem value="csharp">C#</SelectItem>
                <SelectItem value="rust">Rust</SelectItem>
                <SelectItem value="php">PHP</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Framework */}
          <div className="space-y-2">
            <Label htmlFor="framework">Framework</Label>
            <Input
              id="framework"
              value={framework}
              onChange={(e) => setFramework(e.target.value)}
              placeholder="e.g., Express, Django, Spring Boot"
              disabled={submitting}
            />
          </div>

          {/* Target Path */}
          <div className="space-y-2">
            <Label htmlFor="targetPath">Target Path</Label>
            <Input
              id="targetPath"
              value={targetPath}
              onChange={(e) => setTargetPath(e.target.value)}
              placeholder="e.g., /path/to/service"
              disabled={submitting}
            />
          </div>

          {/* Git Repository URL */}
          <div className="space-y-2">
            <Label htmlFor="gitRepoURL">Git Repository URL</Label>
            <Input
              id="gitRepoURL"
              type="url"
              value={gitRepoURL}
              onChange={(e) => setGitRepoURL(e.target.value)}
              placeholder="https://github.com/user/repo.git"
              disabled={submitting}
            />
            <p className="text-xs text-muted-foreground">
              Optional: Connect a git repository if you need coding agents to make changes to the service source code.
            </p>
          </div>

          {/* Error Display */}
          {error && (
            <div className="p-3 text-sm text-red-500 bg-red-500/10 border border-red-500/20 rounded-md">
              {error}
            </div>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={submitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={submitting || !serviceName.trim()}>
              {submitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Adding...
                </>
              ) : (
                "Add Service"
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}


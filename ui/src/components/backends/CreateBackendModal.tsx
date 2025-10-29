import { useState } from "react";
import { useSWRConfig } from "swr";
import { apiClient } from "@/services/api";
import type { CreateBackendRequest } from "@/types/backend";
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
import { Loader2 } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface CreateBackendModalProps {
  isOpen: boolean;
  onClose: () => void;
  onBackendCreated?: (backendId: string) => void;
}

export function CreateBackendModal({
  isOpen,
  onClose,
  onBackendCreated,
}: CreateBackendModalProps) {
  const { mutate } = useSWRConfig();
  const [backendType, setBackendType] = useState("grafana");
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim() || !url.trim()) {
      setError("Please provide a name and URL");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const request: CreateBackendRequest = {
        backend_type: backendType,
        name: name.trim(),
        url: url.trim(),
        username: username.trim() || undefined,
        password: password || undefined,
      };

      const backend = await apiClient.createBackend(request);
      
      // Invalidate backends cache
      mutate("/api/backends");

      // Reset form
      setName("");
      setUrl("");
      setUsername("");
      setPassword("");
      setBackendType("grafana");

      // Notify parent and close modal
      onBackendCreated?.(backend.id);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create backend");
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!submitting) {
      setError(null);
      setName("");
      setUrl("");
      setUsername("");
      setPassword("");
      setBackendType("grafana");
      onClose();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Add Backend</DialogTitle>
          <DialogDescription>
            Connect to an observability backend (Grafana, Prometheus, Jaeger, etc.)
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="backend-type">Backend Type</Label>
            <Select value={backendType} onValueChange={setBackendType} disabled={submitting}>
              <SelectTrigger id="backend-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="grafana">Grafana</SelectItem>
                <SelectItem value="prometheus">Prometheus</SelectItem>
                <SelectItem value="jaeger">Jaeger</SelectItem>
                <SelectItem value="custom">Custom</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Grafana Instance"
              disabled={submitting}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="url">URL</Label>
            <Input
              id="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="http://localhost:3000"
              disabled={submitting}
              required
            />
          </div>

          {(backendType === "grafana" || backendType === "custom") && (
            <>
              <div className="space-y-2">
                <Label htmlFor="username">Username (optional)</Label>
                <Input
                  id="username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="admin"
                  disabled={submitting}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="password">Password (optional)</Label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  disabled={submitting}
                />
              </div>
            </>
          )}

          {error && (
            <div className="p-3 text-sm text-red-500 bg-red-500/10 border border-red-500/20 rounded-md">
              {error}
            </div>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={submitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Add Backend
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}


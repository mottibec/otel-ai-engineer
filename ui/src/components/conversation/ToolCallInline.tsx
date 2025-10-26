import { Badge } from "@/components/ui/badge";
import { Wrench } from "lucide-react";
import { JsonViewer } from "../common/JsonViewer";

interface ToolCallInlineProps {
  toolUse: {
    name: string;
    input: unknown;
  };
}

export function ToolCallInline({ toolUse }: ToolCallInlineProps) {
  return (
    <div className="border-l-2 border-amber-200/40 bg-amber-50/10 dark:bg-amber-950/5 pl-2 py-1 my-1">
      <div className="flex items-center gap-1 mb-0.5">
        <Wrench className="h-2.5 w-2.5 text-amber-500/60 dark:text-amber-400/50 flex-shrink-0" />
        <Badge variant="outline" className="text-[8px] h-3 px-1 font-medium leading-tight border-amber-200/40 text-amber-700/70 dark:text-amber-400/60">
          {toolUse.name}
        </Badge>
      </div>
      <div className="ml-3.5">
        <JsonViewer data={toolUse.input} collapsed title={`Parameters`} />
      </div>
    </div>
  );
}

export type ChatRole = "user" | "agent" | "system";

export interface ToolCallItem {
  id?: string;
  name: string;
  arguments: string;
}

export interface ToolResultItem {
  toolCallId: string;
  action: string;
  result: string;
  error?: string;
  hasError: boolean;
}

export interface ChatMessage {
  id: string;
  role: ChatRole;
  content: string;
  time: number;
  taskId?: string;
  streaming?: boolean;
  reasoning?: string;
  finalAnswer?: string;
  toolCalls?: ToolCallItem[];
  toolResults?: ToolResultItem[];
  audio?: { base64: string; format: string; voice: string; bytes: number };
}

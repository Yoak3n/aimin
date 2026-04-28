export type WsActionType =
  | "Connected"
  | "Log"
  | "Close"
  | "Ping"
  | "Pong"
  | "Interrupt"
  | "Ask"
  | "Answer"
  | "Task"
  | "Reply"
  | "ToolResult"
  | "State";

export interface WsMessage<TAction extends WsActionType = WsActionType, TData = unknown> {
  action: TAction;
  data: TData;
}

export interface WsLogMessageData {
  time: string;
  content: string;
}

export type WsConnectedMessage = WsMessage<"Connected", unknown>;
export type WsLogMessage = WsMessage<"Log", WsLogMessageData>;
export type WsCloseMessage = WsMessage<"Close", "Close" | string>;
export type WsPingMessage = WsMessage<"Ping", "Ping" | string>;
export type WsPongMessage = WsMessage<"Pong", "Pong" | string>;
export type WsInterruptMessage = WsMessage<"Interrupt", "Interrupt" | string | null>;

export interface WsAskMessageData {
  id: string;
  content: string;
}

export interface WsAnswerMessageData {
  id: string;
  content?: string;
  skip?: boolean;
}

export type WsAskMessage = WsMessage<"Ask", WsAskMessageData>;
export type WsAnswerMessage = WsMessage<"Answer", WsAnswerMessageData>;

export type WsTaskType = 0 | 1;

export interface WsTaskData {
  id?: string;
  from: string;
  type: WsTaskType;
  payload: string;
}

export type WsAddTaskMessage = WsMessage<"Task", WsTaskData>;

export type WsReplyStatus = 0 | 1;

export interface WsReplyChunkData {
  task_id: string;
  chunk_idx: number;
  content: string;
}

export interface WsReplyFinishData {
  task_id: string;
  content: string;
}

export interface WsReplyMessageData {
  task_id: string;
  status: WsReplyStatus;
  chunk?: WsReplyChunkData;
  result?: WsReplyFinishData;
}

export type WsReplyMessage = WsMessage<"Reply", WsReplyMessageData>;

export interface WsToolResultMessageData {
  task_id: string;
  tool_call_id: string;
  action: string;
  result: string;
  error?: string;
  has_error: boolean;
}

export type WsToolResultMessage = WsMessage<"ToolResult", WsToolResultMessageData>;
export type WsStateMessage = WsMessage<"State", string>;

export type WsIncomingMessage =
  | WsConnectedMessage
  | WsLogMessage
  | WsCloseMessage
  | WsPingMessage
  | WsPongMessage
  | WsInterruptMessage
  | WsAskMessage
  | WsReplyMessage
  | WsToolResultMessage
  | WsStateMessage;

export type WsOutgoingMessage =
  | WsCloseMessage
  | WsPongMessage
  | WsAnswerMessage
  | WsAddTaskMessage
  | WsPingMessage
  | WsInterruptMessage;

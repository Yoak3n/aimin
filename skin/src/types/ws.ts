export type WsActionType =
  | "Connected"
  | "Log"
  | "Close"
  | "Ping"
  | "Pong"
  | "Ask"
  | "Answer"
  | "Task"
  | "Reply"
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
export type WsStateMessage = WsMessage<"State", string>;

export type WsIncomingMessage =
  | WsConnectedMessage
  | WsLogMessage
  | WsCloseMessage
  | WsPingMessage
  | WsPongMessage
  | WsAskMessage
  | WsReplyMessage
  | WsStateMessage;

export type WsOutgoingMessage = WsCloseMessage | WsPongMessage | WsAnswerMessage | WsAddTaskMessage | WsPingMessage;

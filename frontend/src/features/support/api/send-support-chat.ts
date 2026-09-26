/**
 * send-support-chat.ts — ヘルプチャット送信
 *
 * POST /v1/support/chat
 * マニュアル抜粋（context）を根拠に LLM が回答し、参照記事が sources として返る。
 */
import { useMutation } from "@tanstack/react-query";

import { axios } from "@/lib/axios";

import type {
  SupportChatContextItem,
  SupportChatHistoryMessage,
  SupportChatResponse,
} from "../types";

export interface SendSupportChatParams {
  message: string;
  history: SupportChatHistoryMessage[];
  context: SupportChatContextItem[];
}

async function sendSupportChat(params: SendSupportChatParams): Promise<SupportChatResponse> {
  const { data } = await axios.post<SupportChatResponse>("/v1/support/chat", params);
  return data;
}

export function useSendSupportChat() {
  return useMutation({ mutationFn: sendSupportChat });
}

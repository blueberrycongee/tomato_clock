from __future__ import annotations

import os
import uuid
from typing import Dict, List

from fastapi import FastAPI, HTTPException
from langchain_core.messages import AIMessage, HumanMessage
from pydantic import BaseModel
from fastapi.middleware.cors import CORSMiddleware

from .agent import create_agent

app = FastAPI(title="TomatoClock AI助手 API")

# 允许本地调用
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 在启动时初始化 Agent，仅加载一次，大幅减少推理延迟
agent_executor = create_agent()
# 在内存中简单维护各会话的对话历史（生产环境请替换为持久化会话管理）
conversations: Dict[str, List[HumanMessage | AIMessage]] = {}


class ChatRequest(BaseModel):
    message: str
    session_id: str | None = None


class ChatResponse(BaseModel):
    reply: str


@app.post("/chat", response_model=ChatResponse)
def chat(req: ChatRequest):
    if not req.message.strip():
        raise HTTPException(status_code=400, detail="消息不能为空")

    try:
        reply = agent_executor.invoke({
            "input": req.message,
        })
        reply_str = reply if isinstance(reply, str) else str(reply)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    return ChatResponse(reply=reply_str)


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("agent.server:app", host="0.0.0.0", port=port, reload=False) 
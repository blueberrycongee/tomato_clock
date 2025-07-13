from __future__ import annotations

"""基于 DeepSeek-chat 的自然语言番茄钟助手。

流程：
1. LLM 输出严格 JSON (activity_name, start_time, duration_minutes, label)。
2. 解析 JSON -> ActivityRecord (pydantic)。
3. 调用 log_activity(**record.dict()) 写入本地数据。
"""

import os
from typing import Any, List

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder, HumanMessagePromptTemplate
from langchain_openai import ChatOpenAI

from .output_parser import parse_activity_json, activity_record_parser
from .tools import log_activity
from .utils import current_date_beijing


# ---------------------------------------------------------------------------
# Chain 构造
# ---------------------------------------------------------------------------

def _build_chain():
    """创建完整可调用的 chain。"""

    llm = ChatOpenAI(
        temperature=0,
        model=os.getenv("OPENAI_MODEL", "deepseek-chat"),
        openai_api_base=os.getenv("OPENAI_API_BASE", "https://api.deepseek.com"),
        api_key=os.getenv("OPENAI_API_KEY"),
    )

    current_ts_str = current_date_beijing()

    system_prompt = (
        f"你是一个用于记录用户活动的助手。当前时间（东八区，精确到分钟）为 {current_ts_str}。"  # noqa: E501
        "阅读用户输入后，提取以下字段并以**严格 JSON**输出（不要添加任何解释文字）:"  # noqa: E501
        "activity_name, start_time, duration_minutes, label。"  # noqa: E501
        "要求：\n"
        "1. activity_name 必须来源于用户描述，禁止使用 unknown/未知。\n"
        "2. 将识别到的日期与时间换算为东八区 ISO8601 完整格式 YYYY-MM-DDTHH:MM+08:00，\n"
        "   例如用户说 '2025年7月23日中午12点'，请输出 '2025-07-23T12:00+08:00'。\n"
        "3. 若用户未给出具体日期，假设为今天，并同样输出 ISO8601（包含 +08:00）。\n"
        "4. duration_minutes 必须是正整数。\n"
        "5. label 不确定时填 '自由任务'。\n"
        "6. 若用户使用 '刚刚'、'刚才'、'过去X分钟/小时' 等相对时长描述，必须将 start_time 设置为当前时间减去 duration_minutes，并确保计算出的 (start_time + duration_minutes) 与当前时间相差不超过 5 分钟。\n"
        "7. 若出现时间段关键词且缺少具体时分，默认小时如下：早上08:00、下午14:00、晚上/今晚19:00、凌晨00:30、中午12:00（分钟统一为00）。\n"
        "8. 若用户仅提供日期未提供时分，将 start_time 设为该日期 00:00。\n"
        "9. activity_name 必须是表示活动的中文名词，长度不少于 2 个汉字，不包含 '了' 等时态后缀。\n"
        "   若用户给出的是动词或简称，如 '学'、'吃' 等，请参考以下映射表转为对应名词：\n"
        "   学习: 学, 学了, 读书, 看书\n"
        "   吃饭: 吃饭, 吃, 早餐, 午饭, 晚饭\n"
        "   运动: 运动, 锻炼, 跑步, 健身\n"
        "   工作: 工作, 写代码, coding, 编程\n"
        "   开会: 开会, 会议\n"
        "   睡觉: 睡觉, 午休, 休息\n"
        "   看电视: 看电视, 追剧\n"
        "   看新闻: 看新闻, 刷新闻\n"
        "   社交: 聊天, 交友\n"
        "输出示例：{\"activity_name\": \"锻炼\", \"start_time\": \"2025-07-23T12:00+08:00\", \"duration_minutes\": 30, \"label\": \"健康\"}\n"
        "输出示例：{\"activity_name\": \"学习\", \"start_time\": \"2025-07-23T11:20+08:00\", \"duration_minutes\": 40, \"label\": \"学习\"}\n"
        "输出示例：{\"activity_name\": \"吃饭\", \"start_time\": \"2025-07-23T18:10+08:00\", \"duration_minutes\": 30, \"label\": \"饮食\"}\n\n"
        f"输出格式说明：{activity_record_parser.get_format_instructions()}"
    )

    prompt = ChatPromptTemplate.from_messages(
        [
            SystemMessage(content=system_prompt),
            MessagesPlaceholder(variable_name="chat_history", optional=True),
            HumanMessagePromptTemplate.from_template("{input}"),
        ]
    )

    # ---------------- 调试打印函数 ----------------
    def _llm_with_debug(messages):  # type: ignore[override]
        """在 DEBUG_PROMPT=1 时打印完整消息列表后调用 LLM。"""
        if os.getenv("DEBUG_PROMPT") == "1":
            # 支持 ChatPromptValue 或直接 List[BaseMessage]
            if hasattr(messages, "messages"):
                _msgs = messages.messages  # type: ignore[attr-defined]
            else:
                _msgs = messages  # assume iterable

            print("================ DEBUG PROMPT START ================")
            for msg in _msgs:
                role = msg.__class__.__name__
                content = getattr(msg, "content", str(msg))
                print(f"{role}: {content}\n")
            print("================= DEBUG PROMPT END =================")

        return llm.invoke(messages)

    def _parse_and_call(output: Any) -> str:
        """将 LLM 返回的 AIMessage 或字符串解析为 ActivityRecord 并调用工具"""
        if isinstance(output, AIMessage):
            content = output.content
        else:
            content = str(output)

        record = parse_activity_json(content)
        return log_activity.invoke(record.model_dump())

    return prompt | _llm_with_debug | _parse_and_call


def create_agent():
    """提供给外部调用，返回 chain 对象。"""
    return _build_chain()


# ---------------------------------------------------------------------------
# CLI 入口
# ---------------------------------------------------------------------------

def main():
    print("================ 自然语言番茄钟助手 ================")
    print("提示: 输入 '退出'、'\\q' 或 'quit' 以结束对话\n")

    if not os.getenv("OPENAI_API_KEY"):
        print("[错误] 请在环境变量 OPENAI_API_KEY 中设置 API Key 才能运行。")
        return

    agent = create_agent()

    while True:
        try:
            user_input = input("你: ")
        except (EOFError, KeyboardInterrupt):
            print()
            break

        if user_input.strip().lower() in {"退出", "quit", "\q", "exit"}:
            break

        try:
            reply = agent.invoke({"input": user_input})
            reply_str = reply if isinstance(reply, str) else str(reply)

            print("助手:", reply_str)
        except Exception as err:
            print(f"[错误] {err}")


if __name__ == "__main__":
    main()

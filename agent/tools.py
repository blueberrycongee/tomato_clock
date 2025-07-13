from __future__ import annotations

import json
from datetime import datetime, timedelta
from typing import Optional

import dateparser
from langchain.tools import tool
from langchain_core.pydantic_v1 import BaseModel, Field

from . import utils


class LogActivityInput(BaseModel):
    """定义 log_activity 工具的输入参数结构和描述。"""
    activity_name: str = Field(description="活动的名称，例如 '吃午饭'、'背单词'。")
    start_time: str = Field(description="活动开始时间，推荐使用 ISO8601 或能被 dateparser 解析的中文表达，例如 '2025-07-13T12:20', '中午12点20'。")
    duration_minutes: int = Field(description="活动持续的分钟数，例如 40。")
    label: Optional[str] = Field(description="活动标签，例如 '学习'、'健康'，若无法判断请填 '自由任务'")
# 已移除 date 与 notes 字段，由模型直接决定 label。


@tool(args_schema=LogActivityInput)
def log_activity(
    activity_name: str,
    start_time: str,
    duration_minutes: int,
    label: Optional[str] = None,
) -> str:
    """当用户想要记录任何已经发生的、有持续时长的活动时，调用此工具。
    例如：记录吃饭、学习、开会、运动等。
    """
    try:
        # ---------------- 智能组合时间信息 ----------------
        # 将模型返回的碎片化时间信息组合成更完整的描述，以便 dateparser 更准确地解析
        # DeepSeek 直接给出的表达通常已足够解析，这里仅保留 start_time
        time_parts = [start_time]
        full_time_expr = " ".join(time_parts)
        if not full_time_expr.strip():
            raise ValueError("必须提供有效的开始时间")

        # ---------------- 解析时间 ----------------
        settings = {
            "TIMEZONE": "Asia/Shanghai",
            "TO_TIMEZONE": "Asia/Shanghai",
            "RETURN_AS_TIMEZONE_AWARE": True,
            "PREFER_DATES_FROM": "past",
        }
        base_time = utils.now()
        start_dt: Optional[datetime] = dateparser.parse(
            full_time_expr, languages=["zh"], settings={**settings, "RELATIVE_BASE": base_time}
        )
        if start_dt is None:
            raise ValueError(f"无法解析起始时间表达: {full_time_expr}")

        # 统一精度到分钟
        start_dt = start_dt.replace(second=0, microsecond=0)
        end_dt = start_dt + timedelta(minutes=duration_minutes)
        # dateutil 返回 float，转为 int 确保与 Go 端一致
        actual_duration_sec = int((end_dt - start_dt).total_seconds())

        # ---------------- 读取 & 更新数据 ----------------
        data = utils.load_data()

        task_id: Optional[int] = None
        if activity_name:
            for task in data["tasks"]:
                if task.get("title") == activity_name:
                    task_id = task.get("id")
                    break
            if task_id is None:
                task_id = data["next_task_id"]
                data["next_task_id"] += 1
                now_iso = utils.to_iso(utils.now())
                task_obj = {
                    "id": task_id,
                    "title": activity_name,
                    # 修正：默认标签与 Go 程序保持一致
                    "note": "",
                    "is_done": False,
                    "repeat_rule": "none",
                    "created_at": now_iso,
                    "updated_at": now_iso,
                    "label": label or "自由任务",
                }
                data["tasks"].append(task_obj)

        session_id = data["next_session_id"]
        data["next_session_id"] += 1

        session_obj = {
            "id": session_id,
            "task_id": task_id,
            "mode": "countup",
            "target_seconds": duration_minutes * 60,
            "started_at": utils.to_iso(start_dt),
            "ended_at": utils.to_iso(end_dt),
            "interrupted": False,
            # 修正：使用计算出的实际秒数
            "duration_sec": actual_duration_sec,
        }
        data["sessions"].append(session_obj)

        utils.save_data(data)

        return f"活动 '{activity_name}' 已成功记录。计时ID: {session_id}"
    except Exception as exc:
        return f"记录失败: {exc}" 
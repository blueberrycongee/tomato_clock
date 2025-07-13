from __future__ import annotations

"""自定义输出解析器：将 LLM 生成的 JSON 字符串解析为 ActivityRecord 对象。

DeepSeek-chat 不支持原生 function-calling，因此我们使用 PydanticOutputParser
来验证 / 解析模型输出。当解析成功后，可直接调用 log_activity(**record.dict()).
"""

import json
from typing import Any, Dict

from langchain_core.output_parsers import PydanticOutputParser

# 使用 pydantic v2 原生 BaseModel
from pydantic import BaseModel, Field, ValidationError


class ActivityRecord(BaseModel):
    """LLM 必须生成的 JSON 格式。"""

    activity_name: str = Field(..., description="活动名称")
    start_time: str = Field(..., description="开始时间，可为 ISO8601 或中文表达")
    duration_minutes: int = Field(..., description="持续分钟数")
    label: str | None = Field(None, description="活动标签，如无法判定填写 '自由任务'")


activity_record_parser = PydanticOutputParser(pydantic_object=ActivityRecord)


def parse_activity_json(output: str) -> ActivityRecord:
    """尝试先直接用 PydanticOutputParser 解析；
    若 LLM 返回的不是纯 JSON 而是 ```json ... ``` 包裹，则先去除包裹再解析。"""

    stripped = output.strip()
    if stripped.startswith("```"):
        # 删除围栏与可能的 "json" 标记
        stripped = stripped.strip("`")
        if stripped.startswith("json"):
            stripped = stripped[4:]
        stripped = stripped.strip()
    
    try:
        # 若已是 JSON 字符串
        data: Dict[str, Any] = json.loads(stripped)
    except json.JSONDecodeError:
        # 交给 LangChain PydanticOutputParser，让其报出更友好错误
        try:
            return activity_record_parser.parse(stripped)
        except ValidationError as e:
            raise ValueError(f"输出 JSON 解析失败: {e}")
    else:
        try:
            return ActivityRecord.parse_obj(data)
        except ValidationError as e:
            raise ValueError(f"输出 JSON 字段验证失败: {e}") 
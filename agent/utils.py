from __future__ import annotations

import json
from datetime import datetime, timedelta
from pathlib import Path

import pytz
import requests

# ---------------------------------------------------------------------------
# 依赖
# ---------------------------------------------------------------------------

import os
import tempfile
from typing import Any, Dict

# 新增：跨平台文件锁，避免并发写冲突
import portalocker

# 与 Go 端保持一致的数据文件名
DATA_FILE_NAME = ".tomato_clock.json"

# 北京时区对象
TZ_BEIJING = pytz.timezone("Asia/Shanghai")


# ---------------------------------------------------------------------------
# 文件 / 数据读写
# ---------------------------------------------------------------------------

def _data_file_path() -> Path:
    """返回用户主目录下的数据文件完整路径"""
    return Path.home() / DATA_FILE_NAME


# ---------------------------------------------------------------------------
# JSON 读写
# ---------------------------------------------------------------------------

def load_data() -> Dict[str, Any]:
    """读取 JSON 数据文件。若不存在则返回初始化结构。"""
    path = _data_file_path()
    if not path.exists():
        # 与 Go 端 snake_case 字段保持一致
        return {
            "next_task_id": 1,
            "next_session_id": 1,
            "tasks": [],
            "sessions": [],
        }

    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def save_data(data: Dict[str, Any]) -> None:
    """写回 JSON 数据到文件，使用文件锁和原子替换，防止并发写冲突。"""

    path = _data_file_path()

    # 将数据序列化到临时文件，再原子替换到目标文件
    tmp_fd, tmp_path = tempfile.mkstemp(prefix=".tomato_clock_", suffix=".json")
    try:
        with os.fdopen(tmp_fd, "w", encoding="utf-8") as tmp_file:
            json.dump(data, tmp_file, ensure_ascii=False, indent=2, sort_keys=False)

        # 申请文件锁后再替换，确保跨进程安全
        with portalocker.Lock(str(path) + ".lock", timeout=10):
            os.replace(tmp_path, path)
    finally:
        # 若 os.replace 抛异常，确保临时文件被清理
        if os.path.exists(tmp_path):
            os.remove(tmp_path)


# ---------------------------------------------------------------------------
# 时间工具
# ---------------------------------------------------------------------------

def now() -> datetime:
    """返回带时区的当前时间，时区固定为北京 (UTC+8)。"""
    return datetime.now(TZ_BEIJING)


def to_iso(dt: datetime) -> str:
    """格式化为 ISO 8601 字符串（RFC3339），精确到分钟，保留时区偏移。"""
    dt_local = dt.astimezone(TZ_BEIJING).replace(second=0, microsecond=0)
    try:
        return dt_local.isoformat(timespec="minutes")
    except TypeError:
        # 若运行时 Python 版本不支持 timespec 参数，则退化为手动格式化
        return dt_local.strftime("%Y-%m-%dT%H:%M%z")


def add_minutes(dt: datetime, minutes: int) -> datetime:
    return dt + timedelta(minutes=minutes)


def current_date_beijing() -> str:
    """返回当前时间（东八区）ISO8601 字符串，精确到分钟，例如 2025-07-13T14:25+08:00。

    先尝试在线 WorldTimeAPI；若失败则使用本地时间回退。"""
    try:
        resp = requests.get("https://worldtimeapi.org/api/timezone/Asia/Shanghai", timeout=5)
        if resp.status_code == 200:
            data = resp.json()
            datetime_raw: str = data.get("datetime", "")
            if datetime_raw:
                # WorldTimeAPI 返回形如 "2025-07-13T14:23:31.123456+08:00"
                # 仅保留到分钟
                dt = datetime.fromisoformat(datetime_raw.rstrip("Z"))
                return to_iso(dt)
    except Exception:
        # 网络错误或解析错误时使用本地时间作为回退
        pass
    return to_iso(now()) 
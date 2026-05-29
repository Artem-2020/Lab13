import asyncio
import json
import os
import uuid
from dataclasses import dataclass, field
from typing import Any

from app.logging_config import setup_logger


TASK_SUBJECTS = {
    "checkin_checkout": "hotel.tasks.checkin_checkout",
    "housekeeping": "hotel.tasks.housekeeping",
    "guest_requests": "hotel.tasks.guest_requests",
    "billing": "hotel.tasks.billing",
    "cancellations": "hotel.tasks.cancellations",
}


class TaskTimeoutError(Exception):
    pass


@dataclass
class HotelOrchestrator:
    nats_url: str = field(default_factory=lambda: os.getenv("NATS_URL", "nats://localhost:4222"))
    timeout_seconds: float = 5.0
    max_retries: int = 3
    nc: Any = None
    processed_tasks: int = 0

    def __post_init__(self) -> None:
        self.logger = setup_logger("hotel-orchestrator", "logs/orchestrator.log")

    async def connect(self) -> None:
        if self.nc is None:
            import nats

            self.nc = await nats.connect(self.nats_url)
            self.logger.info("connected to NATS at %s", self.nats_url)

    async def close(self) -> None:
        if self.nc is not None:
            await self.nc.close()
            self.nc = None

    async def run_task(self, task_type: str, payload: dict[str, Any]) -> dict[str, Any]:
        await self.connect()

        if task_type not in TASK_SUBJECTS:
            raise ValueError(f"unsupported task type: {task_type}")

        task_id = payload.get("task_id") or str(uuid.uuid4())
        reply_to = f"hotel.results.{task_id}"
        task = {**payload, "task_id": task_id, "reply_to": reply_to}
        subject = TASK_SUBJECTS[task_type]

        for attempt in range(1, self.max_retries + 1):
            future: asyncio.Future[dict[str, Any]] = asyncio.get_running_loop().create_future()

            async def result_handler(msg: Any) -> None:
                if not future.done():
                    future.set_result(json.loads(msg.data.decode("utf-8")))

            subscription = await self.nc.subscribe(reply_to, cb=result_handler)

            try:
                await self.nc.publish(subject, json.dumps(task).encode("utf-8"))
                self.logger.info("sent task_id=%s type=%s attempt=%d", task_id, task_type, attempt)
                result = await asyncio.wait_for(future, timeout=self.timeout_seconds)
                self.processed_tasks += 1
                self.logger.info("received result task_id=%s total=%d", task_id, self.processed_tasks)
                return result
            except asyncio.TimeoutError:
                self.logger.error("timeout task_id=%s attempt=%d", task_id, attempt)
            finally:
                await subscription.unsubscribe()

        raise TaskTimeoutError(f"task {task_id} timed out after {self.max_retries} attempts")

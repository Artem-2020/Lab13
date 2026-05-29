import asyncio
import json

import pytest

from app.orchestrator import HotelOrchestrator, TaskTimeoutError


class FakeMessage:
    def __init__(self, data: bytes):
        self.data = data


class FakeSubscription:
    async def unsubscribe(self):
        return None


class FakeNATS:
    def __init__(self, should_reply=True):
        self.callbacks = {}
        self.published = []
        self.should_reply = should_reply

    async def subscribe(self, subject, cb):
        self.callbacks[subject] = cb
        return FakeSubscription()

    async def publish(self, subject, payload):
        self.published.append((subject, payload))
        task = json.loads(payload.decode("utf-8"))
        if self.should_reply:
            response = {
                "task_id": task["task_id"],
                "agent_id": "fake-agent",
                "status": "success",
                "message": "ok",
                "room_id": task["room_id"],
                "room_status": "occupied",
            }
            await self.callbacks[task["reply_to"]](FakeMessage(json.dumps(response).encode("utf-8")))

    async def close(self):
        return None


def test_run_task_success():
    asyncio.run(run_success_case())


async def run_success_case():
    orchestrator = HotelOrchestrator(timeout_seconds=0.1)
    orchestrator.nc = FakeNATS()

    result = await orchestrator.run_task(
        "checkin_checkout",
        {
            "action": "checkin",
            "guest_id": "G-1",
            "room_id": "101",
            "reservation_id": "R-1",
            "paid": True,
            "room_ready": True,
        },
    )

    assert result["status"] == "success"
    assert orchestrator.processed_tasks == 1


def test_run_task_timeout_retries_three_times():
    asyncio.run(run_timeout_case())


async def run_timeout_case():
    orchestrator = HotelOrchestrator(timeout_seconds=0.01, max_retries=3)
    orchestrator.nc = FakeNATS(should_reply=False)

    with pytest.raises(TaskTimeoutError):
        await orchestrator.run_task(
            "checkin_checkout",
            {
                "action": "checkin",
                "guest_id": "G-1",
                "room_id": "101",
                "reservation_id": "R-1",
                "paid": True,
                "room_ready": True,
            },
        )

    assert len(orchestrator.nc.published) == 3

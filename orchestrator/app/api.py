from contextlib import asynccontextmanager
from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from app.orchestrator import HotelOrchestrator, TaskTimeoutError


class CheckinCheckoutRequest(BaseModel):
    action: str = Field(pattern="^(checkin|checkout)$")
    guest_id: str
    room_id: str
    reservation_id: str
    paid: bool = False
    room_ready: bool = False


orchestrator = HotelOrchestrator()


@asynccontextmanager
async def lifespan(app: FastAPI):
    await orchestrator.connect()
    yield
    await orchestrator.close()


app = FastAPI(title="Hotel Agent System API", lifespan=lifespan)


@app.post("/tasks/checkin-checkout")
async def create_checkin_checkout_task(request: CheckinCheckoutRequest) -> dict[str, Any]:
    try:
        return await orchestrator.run_task("checkin_checkout", request.model_dump())
    except TaskTimeoutError as exc:
        raise HTTPException(status_code=504, detail=str(exc)) from exc
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc


@app.get("/metrics")
async def metrics() -> dict[str, int]:
    return {"processed_tasks": orchestrator.processed_tasks}


package checkin

import "fmt"

type Task struct {
	TaskID        string `json:"task_id"`
	ReplyTo       string `json:"reply_to"`
	Action        string `json:"action"`
	GuestID       string `json:"guest_id"`
	RoomID        string `json:"room_id"`
	ReservationID string `json:"reservation_id"`
	Paid          bool   `json:"paid"`
	RoomReady     bool   `json:"room_ready"`
}

type Result struct {
	TaskID     string `json:"task_id"`
	AgentID    string `json:"agent_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	RoomID     string `json:"room_id"`
	RoomStatus string `json:"room_status"`
}

func Process(task Task, agentID string) Result {
	if task.TaskID == "" || task.GuestID == "" || task.RoomID == "" || task.ReservationID == "" {
		return failed(task, agentID, "missing required fields")
	}

	switch task.Action {
	case "checkin":
		return processCheckin(task, agentID)
	case "checkout":
		return Result{
			TaskID:     task.TaskID,
			AgentID:    agentID,
			Status:     "success",
			Message:    fmt.Sprintf("guest %s checked out from room %s", task.GuestID, task.RoomID),
			RoomID:     task.RoomID,
			RoomStatus: "needs_cleaning",
		}
	default:
		return failed(task, agentID, "unsupported action")
	}
}

func processCheckin(task Task, agentID string) Result {
	if !task.Paid {
		return failed(task, agentID, "reservation is not paid")
	}
	if !task.RoomReady {
		return failed(task, agentID, "room is not ready")
	}

	return Result{
		TaskID:     task.TaskID,
		AgentID:    agentID,
		Status:     "success",
		Message:    fmt.Sprintf("guest %s checked in to room %s", task.GuestID, task.RoomID),
		RoomID:     task.RoomID,
		RoomStatus: "occupied",
	}
}

func failed(task Task, agentID, message string) Result {
	return Result{
		TaskID:     task.TaskID,
		AgentID:    agentID,
		Status:     "failed",
		Message:    message,
		RoomID:     task.RoomID,
		RoomStatus: "unchanged",
	}
}

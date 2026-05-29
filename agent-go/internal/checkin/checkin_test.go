package checkin

import "testing"

func TestProcessCheckinSuccess(t *testing.T) {
	task := Task{
		TaskID:        "T-1",
		Action:        "checkin",
		GuestID:       "G-1",
		RoomID:        "101",
		ReservationID: "R-1",
		Paid:          true,
		RoomReady:     true,
	}

	result := Process(task, "agent-test")

	if result.Status != "success" {
		t.Fatalf("expected success, got %s", result.Status)
	}
	if result.RoomStatus != "occupied" {
		t.Fatalf("expected occupied, got %s", result.RoomStatus)
	}
}

func TestProcessCheckinFailsWhenRoomIsNotReady(t *testing.T) {
	task := Task{
		TaskID:        "T-2",
		Action:        "checkin",
		GuestID:       "G-2",
		RoomID:        "102",
		ReservationID: "R-2",
		Paid:          true,
		RoomReady:     false,
	}

	result := Process(task, "agent-test")

	if result.Status != "failed" {
		t.Fatalf("expected failed, got %s", result.Status)
	}
}

func TestProcessCheckoutMarksRoomForCleaning(t *testing.T) {
	task := Task{
		TaskID:        "T-3",
		Action:        "checkout",
		GuestID:       "G-3",
		RoomID:        "103",
		ReservationID: "R-3",
	}

	result := Process(task, "agent-test")

	if result.Status != "success" {
		t.Fatalf("expected success, got %s", result.Status)
	}
	if result.RoomStatus != "needs_cleaning" {
		t.Fatalf("expected needs_cleaning, got %s", result.RoomStatus)
	}
}

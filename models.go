package main

import "time"

type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	Tags        []string   `json:"tags,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Stats struct {
	TotalTasks    int            `json:"total_tasks"`
	ByStatus      map[string]int `json:"by_status"`
	OverdueTasks  int            `json:"overdue_tasks"`
	UpcomingTasks int            `json:"upcoming_tasks"`
}

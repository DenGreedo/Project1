package main

import (
	"database/sql"
	"fmt"
	"strings"
)

func CreateTask(db *sql.DB, task Task) (Task, error) {
	tx, err := db.Begin()
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()

	var taskID int
	err = tx.QueryRow(`
        INSERT INTO tasks (title, description, due_date, priority, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at`,
		task.Title, task.Description, task.DueDate, task.Priority, task.Status,
	).Scan(&taskID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return Task{}, err
	}
	task.ID = taskID

	for _, tagName := range task.Tags {
		var tagID int
		err = tx.QueryRow(`SELECT id FROM tags WHERE name = $1`, tagName).Scan(&tagID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow(`INSERT INTO tags (name) VALUES ($1) RETURNING id`, tagName).Scan(&tagID)
			if err != nil {
				return Task{}, err
			}
		} else if err != nil {
			return Task{}, err
		}
		_, err = tx.Exec(`INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2)`, task.ID, tagID)
		if err != nil {
			return Task{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return Task{}, err
	}
	return task, nil
}

func GetTasks(db *sql.DB, filters map[string]string) ([]Task, error) {
	query := `
        SELECT t.id, t.title, t.description, t.due_date, t.priority, t.status, t.created_at, t.updated_at,
               COALESCE(array_agg(tags.name) FILTER (WHERE tags.name IS NOT NULL), '{}') as tags
        FROM tasks t
        LEFT JOIN task_tags tt ON t.id = tt.task_id
        LEFT JOIN tags ON tt.tag_id = tags.id
    `
	var conditions []string
	var args []interface{}
	argIndex := 1

	if status, ok := filters["status"]; ok && status != "" {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}
	if priority, ok := filters["priority"]; ok && priority != "" {
		conditions = append(conditions, fmt.Sprintf("t.priority = $%d", argIndex))
		args = append(args, priority)
		argIndex++
	}
	if tag, ok := filters["tag"]; ok && tag != "" {
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM task_tags tt2 JOIN tags tg2 ON tt2.tag_id = tg2.id WHERE tt2.task_id = t.id AND tg2.name = $%d)",
			argIndex))
		args = append(args, tag)
		argIndex++
	}
	if search, ok := filters["search"]; ok && search != "" {
		conditions = append(conditions, fmt.Sprintf("(t.title ILIKE $%d OR t.description ILIKE $%d)", argIndex, argIndex+1))
		args = append(args, "%"+search+"%", "%"+search+"%")
		argIndex += 2
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " GROUP BY t.id"

	sortBy := filters["sort_by"]
	if sortBy == "" {
		sortBy = "created_at"
	}
	order := filters["order"]
	if order == "" {
		order = "DESC"
	}
	allowedSortFields := map[string]bool{
		"id": true, "title": true, "due_date": true,
		"priority": true, "status": true, "created_at": true, "updated_at": true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "created_at"
	}
	if strings.ToUpper(order) != "ASC" && strings.ToUpper(order) != "DESC" {
		order = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, strings.ToUpper(order))

	limit := 10
	offset := 0
	if l, ok := filters["limit"]; ok && l != "" {
		fmt.Sscanf(l, "%d", &limit)
		if limit <= 0 || limit > 100 {
			limit = 10
		}
	}
	if o, ok := filters["offset"]; ok && o != "" {
		fmt.Sscanf(o, "%d", &offset)
		if offset < 0 {
			offset = 0
		}
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var dueDate sql.NullTime
		var tags []string
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &dueDate,
			&task.Priority, &task.Status, &task.CreatedAt, &task.UpdatedAt, &tags)
		if err != nil {
			return nil, err
		}
		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}
		task.Tags = tags
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func GetTask(db *sql.DB, id int) (Task, error) {
	var task Task
	var dueDate sql.NullTime
	var tags []string
	err := db.QueryRow(`
        SELECT t.id, t.title, t.description, t.due_date, t.priority, t.status, t.created_at, t.updated_at,
               COALESCE(array_agg(tags.name) FILTER (WHERE tags.name IS NOT NULL), '{}') as tags
        FROM tasks t
        LEFT JOIN task_tags tt ON t.id = tt.task_id
        LEFT JOIN tags ON tt.tag_id = tags.id
        WHERE t.id = $1
        GROUP BY t.id`, id,
	).Scan(&task.ID, &task.Title, &task.Description, &dueDate,
		&task.Priority, &task.Status, &task.CreatedAt, &task.UpdatedAt, &tags)
	if err != nil {
		return Task{}, err
	}
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}
	task.Tags = tags
	return task, nil
}

func UpdateTask(db *sql.DB, id int, task Task) (Task, error) {
	tx, err := db.Begin()
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
        UPDATE tasks SET title=$1, description=$2, due_date=$3, priority=$4, status=$5, updated_at=NOW()
        WHERE id=$6`,
		task.Title, task.Description, task.DueDate, task.Priority, task.Status, id)
	if err != nil {
		return Task{}, err
	}

	_, err = tx.Exec(`DELETE FROM task_tags WHERE task_id=$1`, id)
	if err != nil {
		return Task{}, err
	}
	for _, tagName := range task.Tags {
		var tagID int
		err = tx.QueryRow(`SELECT id FROM tags WHERE name = $1`, tagName).Scan(&tagID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow(`INSERT INTO tags (name) VALUES ($1) RETURNING id`, tagName).Scan(&tagID)
			if err != nil {
				return Task{}, err
			}
		} else if err != nil {
			return Task{}, err
		}
		_, err = tx.Exec(`INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2)`, id, tagID)
		if err != nil {
			return Task{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return Task{}, err
	}

	return GetTask(db, id)
}

func DeleteTask(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM tasks WHERE id=$1`, id)
	return err
}

func CreateTag(db *sql.DB, name string) (Tag, error) {
	var tag Tag
	err := db.QueryRow(`INSERT INTO tags (name) VALUES ($1) RETURNING id, name`, name).Scan(&tag.ID, &tag.Name)
	if err != nil {
		return Tag{}, err
	}
	return tag, nil
}

func GetTags(db *sql.DB) ([]Tag, error) {
	rows, err := db.Query(`SELECT id, name FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func GetStats(db *sql.DB) (Stats, error) {
	stats := Stats{ByStatus: make(map[string]int)}

	err := db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&stats.TotalTasks)
	if err != nil {
		return stats, err
	}

	rows, err := db.Query(`SELECT status, COUNT(*) FROM tasks GROUP BY status`)
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return stats, err
		}
		stats.ByStatus[status] = count
	}

	err = db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE due_date IS NOT NULL AND due_date < NOW() AND status != 'done'`).Scan(&stats.OverdueTasks)
	if err != nil {
		return stats, err
	}
	err = db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE due_date IS NOT NULL AND due_date >= NOW() AND status != 'done'`).Scan(&stats.UpcomingTasks)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

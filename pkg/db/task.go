package db

import (
	"database/sql"
	"fmt"

)

type Task struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	query := `
		INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (?, ?, ?, ?)
	`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("db exec error: %w", err)
	}

	return res.LastInsertId()
}

func GetTasks(limit int) ([]*Task, error) {
	query := `
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		ORDER BY date ASC 
		LIMIT ?
	`

	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(
			&task.ID,
			&task.Date,
			&task.Title,
			&task.Comment,
			&task.Repeat,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, nil
}

func GetTask(id int64) (*Task, error) {
	query := `
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE id = ?
    `

	var task Task
	err := DB.QueryRow(query, id).Scan(
		&task.ID,
		&task.Date,
		&task.Title,
		&task.Comment,
		&task.Repeat,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &task, nil
}

func UpdateTask(task *Task) error {
    query := `
        UPDATE scheduler 
        SET date = ?, title = ?, comment = ?, repeat = ?
        WHERE id = ?
    `
    
    res, err := DB.Exec(query, 
        task.Date,
        task.Title,
        task.Comment,
        task.Repeat,
        task.ID,
    )
    
    if err != nil {
        return fmt.Errorf("exec error: %w", err)
    }
    
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("rows affected error: %w", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("task not found")
    }
    
    return nil
}

func DeleteTask(id int64) error {
    query := `DELETE FROM scheduler WHERE id = ?`
    res, err := DB.Exec(query, id)
    if err != nil {
        return fmt.Errorf("delete error: %w", err)
    }

    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("rows affected error: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("task not found")
    }

    return nil
}

func UpdateTaskDate(id int64, newDate string) error {
    query := `UPDATE scheduler SET date = ? WHERE id = ?`
    res, err := DB.Exec(query, newDate, id)
    if err != nil {
        return fmt.Errorf("update date error: %w", err)
    }

    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("rows affected error: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("task not found")
    }

    return nil
}
package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

type List struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	ItemCount int
	DoneCount int
}

type Item struct {
	ID        int64
	ListID    int64
	Text      string
	Done      bool
	CreatedAt time.Time
}

func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *DB) migrate() error {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS lists (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT    NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS items (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			list_id    INTEGER NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
			text       TEXT    NOT NULL,
			done       INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

// Lists

func (d *DB) GetLists() ([]List, error) {
	rows, err := d.conn.Query(`
		SELECT l.id, l.name, l.created_at,
		       COUNT(i.id) as item_count,
		       SUM(CASE WHEN i.done=1 THEN 1 ELSE 0 END) as done_count
		FROM lists l
		LEFT JOIN items i ON i.list_id = l.id
		GROUP BY l.id
		ORDER BY l.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []List
	for rows.Next() {
		var l List
		var doneCount sql.NullInt64
		if err := rows.Scan(&l.ID, &l.Name, &l.CreatedAt, &l.ItemCount, &doneCount); err != nil {
			return nil, err
		}
		l.DoneCount = int(doneCount.Int64)
		lists = append(lists, l)
	}
	return lists, nil
}

func (d *DB) GetList(id int64) (*List, error) {
	var l List
	var doneCount sql.NullInt64
	err := d.conn.QueryRow(`
		SELECT l.id, l.name, l.created_at,
		       COUNT(i.id) as item_count,
		       SUM(CASE WHEN i.done=1 THEN 1 ELSE 0 END) as done_count
		FROM lists l
		LEFT JOIN items i ON i.list_id = l.id
		WHERE l.id = ?
		GROUP BY l.id
	`, id).Scan(&l.ID, &l.Name, &l.CreatedAt, &l.ItemCount, &doneCount)
	if err != nil {
		return nil, err
	}
	l.DoneCount = int(doneCount.Int64)
	return &l, nil
}

func (d *DB) CreateList(name string) (*List, error) {
	res, err := d.conn.Exec(`INSERT INTO lists (name) VALUES (?)`, name)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetList(id)
}

func (d *DB) RenameList(id int64, name string) error {
	_, err := d.conn.Exec(`UPDATE lists SET name=? WHERE id=?`, name, id)
	return err
}

func (d *DB) DeleteList(id int64) error {
	_, err := d.conn.Exec(`DELETE FROM lists WHERE id=?`, id)
	return err
}

// Items

func (d *DB) GetItems(listID int64) ([]Item, error) {
	rows, err := d.conn.Query(`
		SELECT id, list_id, text, done, created_at
		FROM items WHERE list_id=?
		ORDER BY done ASC, text ASC
	`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var it Item
		var done int
		if err := rows.Scan(&it.ID, &it.ListID, &it.Text, &done, &it.CreatedAt); err != nil {
			return nil, err
		}
		it.Done = done == 1
		items = append(items, it)
	}
	return items, nil
}

func (d *DB) CreateItem(listID int64, text string) (*Item, error) {
	res, err := d.conn.Exec(`INSERT INTO items (list_id, text) VALUES (?, ?)`, listID, text)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	var it Item
	var done int
	err = d.conn.QueryRow(`SELECT id, list_id, text, done, created_at FROM items WHERE id=?`, id).
		Scan(&it.ID, &it.ListID, &it.Text, &done, &it.CreatedAt)
	it.Done = done == 1
	return &it, err
}

func (d *DB) ToggleItem(id int64) (*Item, error) {
	_, err := d.conn.Exec(`UPDATE items SET done = CASE WHEN done=1 THEN 0 ELSE 1 END WHERE id=?`, id)
	if err != nil {
		return nil, err
	}
	var it Item
	var done int
	err = d.conn.QueryRow(`SELECT id, list_id, text, done, created_at FROM items WHERE id=?`, id).
		Scan(&it.ID, &it.ListID, &it.Text, &done, &it.CreatedAt)
	it.Done = done == 1
	return &it, err
}

func (d *DB) UpdateItem(id int64, text string) (*Item, error) {
	_, err := d.conn.Exec(`UPDATE items SET text=? WHERE id=?`, text, id)
	if err != nil {
		return nil, err
	}
	var it Item
	var done int
	err = d.conn.QueryRow(`SELECT id, list_id, text, done, created_at FROM items WHERE id=?`, id).
		Scan(&it.ID, &it.ListID, &it.Text, &done, &it.CreatedAt)
	it.Done = done == 1
	return &it, err
}

func (d *DB) DeleteItem(id int64) error {
	_, err := d.conn.Exec(`DELETE FROM items WHERE id=?`, id)
	return err
}

func (d *DB) ClearChecked(listID int64) error {
	_, err := d.conn.Exec(`DELETE FROM items WHERE list_id=? AND done=1`, listID)
	return err
}

func (d *DB) CheckAll(listID int64) error {
	_, err := d.conn.Exec(`UPDATE items SET done=1 WHERE list_id=?`, listID)
	return err
}
func (d *DB) UncheckAll(listID int64) error {
	_, err := d.conn.Exec(`UPDATE items SET done=0 WHERE list_id=?`, listID)
	return err
}
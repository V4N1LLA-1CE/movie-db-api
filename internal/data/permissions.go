package data

import (
	"context"
	"database/sql"
	"slices"
	"time"
)

type Permissions []string

func (p *Permissions) Include(code string) bool {
	return slices.Contains(*p, code)
}

type PermissionModel struct {
	DB *sql.DB
}

func (m *PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	stmt := `SELECT permissions.code
  FROM permissions
  INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
  INNER JOIN users ON users_permissions.user_id = user.id
  WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var p string

		err := rows.Scan(&p)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

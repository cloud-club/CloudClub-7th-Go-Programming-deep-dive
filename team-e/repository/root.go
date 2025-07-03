package repository

import (
	"database/sql"
	"feather/config"
	"feather/types"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

type Repository struct {
	config *config.Config
	db     *sql.DB
}

const (
	users     = "feather.user"
	basecamps = "feather.basecamp"
	projects  = "feather.project"
)

func NewRepository(c *config.Config) (*Repository, error) {
	r := &Repository{config: c}
	var err error

	if r.db, err = sql.Open(c.DB.Database, c.DB.URL); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repository) CreateUser(email string, password string, nickname string) error {
	if _, err := r.db.Exec("INSERT INTO feather.user(email, password, nickname) VALUES(?, ?, ?)",
		email, password, nickname); err != nil {
		return err
	}
	log.Println("CreateUser Query run successfully!")
	return nil
}

func (r *Repository) User(userId int64) (*types.User, error) {
	u := new(types.User)
	qs := query([]string{"SELECT id, email, nickname FROM", users, "WHERE id = ?"})
	if err := r.db.QueryRow(qs, userId).Scan(&u.ID, &u.Email, &u.Nickname); err != nil {
		if err := noResult(err); err != nil {
			return nil, err
		}
	}

	log.Println("User Query run successfully!")
	return u, nil
}

func (r *Repository) CreateBasecamp(name string, url string, token string, owner string, userId int64) error {
	if _, err := r.db.Exec("INSERT INTO feather.basecamp(name, url, token, owner, user_id) VALUES(?, ?, ?, ?, ?)",
		name, url, token, owner, userId); err != nil {
		return err
	}
	log.Println("CreateBasecamp Query run successfully!")
	return nil
}

func (r *Repository) TokenByBasecampId(baseCampId int64) (string, error) {
	var token string
	qs := query([]string{"SELECT token FROM", basecamps, "WHERE id = ?"})

	if err := r.db.QueryRow(qs, baseCampId).Scan(&token); err != nil {
		if err := noResult(err); err != nil {
			return "", err
		}
	}

	log.Println("TokenByBasecampId Query run successfully!")
	return token, nil
}

func (r *Repository) BasecampsByUserId(userId int64) ([]*types.Basecamp, error) {
	qs := query([]string{"SELECT id, name, url, owner, token, user_id FROM", basecamps, "WHERE user_id = ?"})
	rows, err := r.db.Query(qs, userId)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	var baseCamps []*types.Basecamp

	for rows.Next() {
		b := new(types.Basecamp)
		if err := rows.Scan(&b.ID, &b.Name, &b.URL, &b.Owner, &b.Token, &b.User_ID); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		baseCamps = append(baseCamps, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Found %d basecamps for user_id=%d\n", len(baseCamps), userId)
	return baseCamps, nil
}

func (r *Repository) Basecamp(baseCampId int64) (*types.Basecamp, error) {
	b := new(types.Basecamp)
	qs := query([]string{"SELECT * FROM", basecamps, "WHERE id = ?"})

	if err := r.db.QueryRow(qs, baseCampId).Scan(&b.ID, &b.Name, &b.URL, &b.Owner, &b.Token, &b.User_ID); err != nil {
		if err := noResult(err); err != nil {
			return nil, err
		}
	}

	log.Println("BaseCamp Query run successfully!")
	return b, nil
}

func (r *Repository) CreateProject(name string, url string, owner string, private bool, baseCampId int64) error {
	if _, err := r.db.Exec("INSERT INTO feather.project(name, url, owner, private, basecamp_id) VALUES(?, ?, ?, ?, ?)",
		name, url, owner, private, baseCampId); err != nil {
		return err
	}
	log.Println("CreateProject Query run successfully!")
	return nil
}

func (r *Repository) ProjectsByBaseCampId(baseCampId int64) ([]*types.Project, error) {
	qs := query([]string{"SELECT * FROM", projects, "WHERE basecamp_id = ?"})
	rows, err := r.db.Query(qs, baseCampId)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	var projects []*types.Project

	for rows.Next() {
		p := new(types.Project)
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.Owner, &p.Private, &p.Basecamp_ID); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Found %d projects for basecamp_id=%d\n", len(projects), baseCampId)
	return projects, nil
}

func (r *Repository) Project(projectId int64) (*types.Project, error) {
	p := new(types.Project)
	qs := query([]string{"SELECT * FROM", projects, "WHERE id = ?"})

	if err := r.db.QueryRow(qs, projectId).Scan(&p.ID, &p.Name, &p.URL, &p.Owner, &p.Private, &p.Basecamp_ID); err != nil {
		if err := noResult(err); err != nil {
			return nil, err
		}
	}

	log.Println("Project Query run successfully!")
	return p, nil
}

func (r *Repository) ProjectWithBaseCampInfo(projectId int64) (*types.ProjectWithBaseCampInfo, error) {
	pb := new(types.ProjectWithBaseCampInfo)
	qs := query([]string{
		"SELECT",
		"p.id AS project_id,",        // Project ID
		"p.name AS project_name,",    // Project Name
		"p.url AS project_url,",      // Project URL
		"p.owner AS project_owner,",  // Project Owner
		"b.name AS basecamp_name,",   // BaseCamp Name
		"b.url AS basecamp_url,",     // BaseCamp URL
		"b.owner AS basecamp_owner,", // BaseCamp Owner
		"b.token AS token",           // Token
		"FROM", projects, "p",        // Alias 'projects' table as 'p'
		"JOIN", basecamps, "b", // Alias 'base_camps' table as 'b'
		"ON p.basecamp_id = b.id", // JOIN condition
		"WHERE p.id = ?",
	})
	if err := r.db.QueryRow(qs, projectId).Scan(
		&pb.ProjectID,
		&pb.ProjectName,
		&pb.ProjectURL,
		&pb.ProjectOwner,
		&pb.BaseCampName,
		&pb.BaseCampURL,
		&pb.BaseCampOwner,
		&pb.Token,
	); err != nil {
		if err := noResult(err); err != nil {
			return nil, err
		}
	}

	log.Println("Project Query run successfully!")
	return pb, nil
}

func query(qs []string) string {
	return strings.Join(qs, " ") + ";"
}

func noResult(err error) error {
	if strings.Contains(err.Error(), "sql: no rows in result set") {
		return nil
	} else {
		return err
	}
}

package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var (
	DB *sql.DB
)

const (
	host         = "localhost"
	port         = "5433"
	user         = "postgres"
	databaseName = "blog_test"
	sslMode      = "disable"
)

func init() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, "password", databaseName, sslMode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	DB = db
}

type Blog struct {
	Id        string
	Author_Id string
	Title     string
	Content   string
}

func CreateBlog(id string, authorId string, title string, content string) (*Blog, error) {
	statement := `
	INSERT INTO blog(id, author_id, title, content) 
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`
	var resId string
	err := DB.QueryRow(statement, id, authorId, title, content).Scan(&resId)
	if err != nil {
		return nil, err
	}
	return &Blog{
		Id: resId,
	}, nil
}

func GetBlog(id string) (*Blog, error) {
	statement := `
	SELECT * FROM blog WHERE id=$1 LIMIT 1
	`
	var blog Blog
	err := DB.QueryRow(statement, id).Scan(&blog.Id, &blog.Author_Id, &blog.Title, &blog.Content)
	if err != nil {
		return nil, err
	}
	return &blog, nil
}

func ListBlogs() ([]*Blog, error) {
	statement := `
		SELECT * FROM blog
	`

	rows, err := DB.Query(statement)
	if err != nil {
		return nil, err
	}
	var blog Blog
	var res []*Blog

	defer rows.Close()
	for rows.Next() {
		rows.Scan(
			&blog.Id,
			&blog.Author_Id,
			&blog.Title,
			&blog.Content,
		)
		res = append(res, &blog)
	}
	return res, nil
}

func UpdateBlog(id string, author_id string, title string, content string) (*Blog, error) {
	statement := `
		UPDATE blog SET author_id = $1, title = $2, content = $3 WHERE id = $4 RETURNING id, author_id, title, content;
	`
	var blog Blog
	err := DB.QueryRow(statement, author_id, title, content, id).Scan(&blog.Id, &blog.Author_Id, &blog.Title, &blog.Content)
	if err != nil {
		return nil, err
	}

	return &blog, nil
}

func DeleteBlog(id string) error {
	statement := `
		DELETE FROM blog WHERE id = $1 RETURNING id;
	`
	DB.QueryRow(statement, id)

	return nil
}

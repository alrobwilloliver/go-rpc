package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-grpc-course/blog/blogpb"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	_ "modernc.org/sqlite"
)

type executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

var store *Store

type Store struct {
	// this is the actual connection(pool) to the db which has the Begin() method
	db       *sql.DB
	executor executor
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	_, execErr := db.Exec("CREATE TABLE IF NOT EXISTS blog(author_id TEXT, title TEXT, content TEXT, id TEXT, PRIMARY KEY(id), UNIQUE(id));")
	if execErr != nil {
		log.Fatalf("there was an error creating the blog table: %v", execErr)
	}
	// the initial store contains just the connection(pool)
	return &Store{db, db}, nil
}

type blogItem struct {
	ID       string
	AuthorId string
	Content  string
	Title    string
}

type server struct{}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	blog := req.GetBlog()

	data := blogItem{
		ID:       blog.Id,
		AuthorId: blog.AuthorId,
		Content:  blog.Content,
		Title:    blog.Title,
	}

	_, err := store.executor.Exec("INSERT INTO blog(id, author_id, content, title) VALUES (?, ?, ?, ?);", data.ID, data.AuthorId, data.Content, data.Title)
	if err != nil {
		log.Fatalf("Failed to add blog content to the database: %v", err)
		return nil, err
	}
	return &blogpb.CreateBlogResponse{
		Blog: blog,
	}, nil
}

func (*server) GetBlog(ctx context.Context, req *blogpb.GetBlogRequest) (*blogpb.GetBlogResponse, error) {
	id := req.GetId()

	row, execErr := store.executor.Query("SELECT id, author_id, title, content FROM blog WHERE id = ?;", id)
	if execErr != nil {
		log.Fatalf("Failed to get the blog of id: %s, with error: %v", id, execErr)
	}
	defer row.Close()
	var blog blogpb.Blog
	for row.Next() {
		scanErr := row.Scan(
			&blog.Id,
			&blog.AuthorId,
			&blog.Title,
			&blog.Content,
		)
		if scanErr != nil {
			log.Fatalf("Error marshalling GetBlog response: %v", scanErr)
		}
	}

	return &blogpb.GetBlogResponse{
		Blog: &blog,
	}, nil
}

func (*server) ListBlogs(ctx context.Context, req *blogpb.ListBlogRequest) (*blogpb.ListBlogResponse, error) {
	rows, err := store.executor.Query("SELECT id, author_id, title, content FROM blog;")
	if err != nil {
		log.Fatalf("Error querying for ListBlogs in the database: %v", err)
		return nil, err
	}
	var blog blogpb.Blog
	var res []*blogpb.Blog

	defer rows.Close()
	for rows.Next() {
		rows.Scan(
			&blog.Id,
			&blog.AuthorId,
			&blog.Title,
			&blog.Content,
		)
		res = append(res, &blog)
	}

	return &blogpb.ListBlogResponse{
		Blog: res,
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	reqBlog := req.GetBlog()
	fmt.Printf("reqBlog: %v", reqBlog)

	row, execErr := store.executor.Query("UPDATE blog SET author_id = ?, title = ?, content = ? WHERE id = ? RETURNING id, author_id, title, content;", reqBlog.AuthorId, reqBlog.Title, reqBlog.Content, reqBlog.Id)
	if execErr != nil {
		log.Fatalf("Failed to get the blog of id: %s, with error: %v", reqBlog.Id, execErr)
	}
	defer row.Close()
	var blog blogpb.Blog
	for row.Next() {
		scanErr := row.Scan(
			&blog.Id,
			&blog.AuthorId,
			&blog.Title,
			&blog.Content,
		)
		if scanErr != nil {
			log.Fatalf("Error marshalling GetBlog response: %v", scanErr)
		}
	}

	return &blogpb.UpdateBlogResponse{
		Blog: &blog,
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	id := req.GetId()

	row, queryErr := store.executor.Query("DELETE FROM blog WHERE id = ?;", id)
	if queryErr != nil {
		log.Fatalf("Failed to delete the blog of id: %s, with error: %v", id, queryErr)
		return nil, queryErr
	}
	defer row.Close()
	var blog blogpb.Blog
	for row.Next() {
		scanErr := row.Scan(
			&blog.Id,
			&blog.AuthorId,
			&blog.Title,
			&blog.Content,
		)
		if scanErr != nil {
			log.Fatalf("Error marshalling GetBlog response: %v", scanErr)
			return nil, scanErr
		}
	}

	fmt.Printf("Deleted the blog of id: %v", id)

	return &blogpb.DeleteBlogResponse{}, nil
}

func main() {
	fmt.Println("Starting RPC Blog Server...")

	dsn := "file:/home/alrob/personalGo/go-grpc-course/blog/store/data.db"
	var err error
	store, err = NewStore(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database %v", err)
		return
	}

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
		return
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for control c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// block until a signal is received
	<-ch
	fmt.Println("Closing the db")
	store.db.Close()
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Close the listener")
	lis.Close()
	fmt.Println("End of Program")
}

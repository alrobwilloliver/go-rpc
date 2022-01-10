package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-grpc-course/blog/blogpb"
	db "go-grpc-course/store"
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

	_, err := db.CreateBlog(blog.Id, blog.AuthorId, blog.Title, blog.Content)
	if err != nil {
		return nil, err
	}
	blogRes := blogpb.Blog{
		Id: blog.Id,
	}
	return &blogpb.CreateBlogResponse{
		Blog: &blogRes,
	}, nil
}

func (*server) GetBlog(ctx context.Context, req *blogpb.GetBlogRequest) (*blogpb.GetBlogResponse, error) {
	id := req.GetId()

	res, err := db.GetBlog(id)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Server blog %v", res)

	blog := &blogpb.Blog{
		Id:       res.Id,
		AuthorId: res.Author_Id,
		Content:  res.Content,
		Title:    res.Title,
	}

	return &blogpb.GetBlogResponse{
		Blog: blog,
	}, nil
}

func (*server) ListBlogs(ctx context.Context, req *blogpb.ListBlogRequest) (*blogpb.ListBlogResponse, error) {

	res, err := db.ListBlogs()
	if err != nil {
		return nil, err
	}
	var blogs []*blogpb.Blog
	for _, blog := range res {
		blogpb := &blogpb.Blog{
			Id:       blog.Id,
			AuthorId: blog.Author_Id,
			Content:  blog.Content,
			Title:    blog.Title,
		}
		blogs = append(blogs, blogpb)
	}

	return &blogpb.ListBlogResponse{
		Blog: blogs,
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	reqBlog := req.GetBlog()

	res, err := db.UpdateBlog(reqBlog.Id, reqBlog.AuthorId, reqBlog.Title, reqBlog.Content)
	if err != nil {
		return nil, err
	}
	blog := &blogpb.Blog{
		Id:       res.Id,
		AuthorId: res.Author_Id,
		Content:  res.Content,
		Title:    res.Title,
	}

	return &blogpb.UpdateBlogResponse{
		Blog: blog,
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	id := req.GetId()

	err := db.DeleteBlog(id)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Deleted blog with id: %s", id)

	return &blogpb.DeleteBlogResponse{}, nil
}

func main() {

	// var (
	// 	host         string
	// 	port         string
	// 	user         string
	// 	databaseName string
	// 	password     string
	// 	sslMode      string
	// )

	// fs := flag.NewFlagSet("blog", flag.ExitOnError)
	// fs.StringVar(&host, "host", "localhost", "the host for the database")
	// fs.StringVar(&port, "port", "5433", "port to host the database on")
	// fs.StringVar(&user, "postgres-user", "postgres", "postgres database username")
	// fs.StringVar(&databaseName, "databaseName", "blog_test", "database name")
	// fs.StringVar(&password, "password", "", "database password")
	// fs.StringVar(&sslMode, "sslMode", "disabled", "set ssl mode enabled or disabled")

	fmt.Println("Starting RPC Blog Server...")

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
	db.DB.Close()
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Close the listener")
	lis.Close()
	fmt.Println("End of Program")
}

package main

import (
	"context"
	"fmt"
	"go-grpc-course/blog/blogpb"
	"log"

	"google.golang.org/grpc"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)

	blog := &blogpb.Blog{
		// Id:       uuid.New().String(),
		Id:       "0543076c-9666-4630-addc-014ddb214ded",
		AuthorId: "Alan Oliver",
		Content:  "Content of blog post edited 1",
		Title:    "Title of Stuff edited 1",
	}

	createRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
		Blog: blog,
	})
	if err != nil {
		log.Fatal("there's a problem here %v", err)
	}
	fmt.Printf("Response: %v", createRes)

	// updateRes, err := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: blog})
	// if err != nil {
	// 	log.Fatal("there's a problem here: %v", err)
	// }
	// fmt.Printf("Response: %v", updateRes)

	// listRes, err := c.ListBlogs(context.Background(), &blogpb.ListBlogRequest{})
	// if err != nil {
	// 	log.Fatal("there's a problem here: %v", err)
	// }
	// fmt.Printf("Response: %v", listRes)

	getRes, err := c.GetBlog(context.Background(), &blogpb.GetBlogRequest{Id: "0543076c-9666-4630-addc-014ddb214ded"})
	if err != nil {
		log.Fatal("there's a problem here: %v", err)
	}
	fmt.Printf("Response: %v", getRes)

	// delRes, err := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{Id: "5aee4bb4-7514-43dd-9593-4533a044449b"})
	// if err != nil {
	// 	log.Fatal("there's a problem here: %v", err)
	// }
	// fmt.Printf("Response: %v", delRes)
}

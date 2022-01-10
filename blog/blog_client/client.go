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

	// blog := &blogpb.Blog{
	// 	Id: uuid.New().String(),
	// 	// Id:       "d09f1d59-ce07-421b-9804-fa66b8c4da08",
	// 	AuthorId: "Alan Oliver",
	// 	Content:  "Content of blog post for list 3",
	// 	Title:    "Title of Stuff for list 3",
	// }

	// createRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
	// 	Blog: blog,
	// })
	// if err != nil {
	// 	log.Fatal("there's a problem here %v", err)
	// }
	// fmt.Printf("Response: %v", createRes)

	// updateRes, err := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: blog})
	// if err != nil {
	// 	log.Fatal("there's a problem here: %v", err)
	// }
	// fmt.Printf("Response: %v", updateRes)

	listRes, err := c.ListBlogs(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatal("there's a problem here: %v", err)
	}
	fmt.Printf("Response: %v \n", listRes)

	// getRes, err := c.GetBlog(context.Background(), &blogpb.GetBlogRequest{Id: "d09f1d59-ce07-421b-9804-fa66b8c4da08"})
	// if err != nil {
	// 	log.Fatal("there's a problem here: %v", err)
	// }
	// fmt.Printf("Response: %v \n", getRes)

	// delRes, err := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{Id: "d09f1d59-ce07-421b-9804-fa66b8c4da08"})
	// if err != nil {
	// 	log.Fatal("there's a problem here: %v", err)
	// }
	// fmt.Printf("Response: %v", delRes)
}

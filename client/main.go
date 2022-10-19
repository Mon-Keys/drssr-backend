package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	pb "client/recognize"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RecognizeAPIClient interface {
	RecognizePhoto(ctx context.Context, image []byte) (string, error)
}

type recognizeAPIClient struct {
	classMap map[string][]int64
	conn     *grpc.ClientConn
	client   pb.RecognizeApiServiceClient
}

func (r *recognizeAPIClient) RecognizePhoto(ctx context.Context, image []byte) (string, error) {
	res, err := r.client.RecognizePhoto(ctx, &pb.RecognizePhotoRequest{
		Image: image,
	})
	if err != nil {
		return "", err
	}

	return res.Category, nil
}

func NewRecognizeApiClient(address string) (RecognizeAPIClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := pb.NewRecognizeApiServiceClient(conn)

	client := &recognizeAPIClient{
		classMap: map[string][]int64{
			"Hoodie":  {6, 5},
			"Skirt":   {9},
			"Tee":     {8, 7},
			"Sweater": {6, 5},
			"Jacket":  {13, 14},
			"Dress":   {10},
		},
		conn:   conn,
		client: c,
	}
	return client, nil
}

func main() {
	conn, err := NewRecognizeApiClient("localhost:8082")
	if err != nil {
		log.Fatal("Failed to connect to recognizeAPI: ", err)
	}

	fileBytes, err := ioutil.ReadFile("hoodie.jpg")
	if err != nil {
		panic(err)
	}

	lol, err := conn.RecognizePhoto(context.Background(), fileBytes)
	if err != nil {
		panic(err)
	}

	fmt.Println(lol)
}

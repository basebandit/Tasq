package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	v1 "github.com/basebandit/go-grpc/pkg/api/v1"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
)

const (
	//apiVersion is the version of API supported by the server
	apiVersion = "v1"
)

func main() {
	//get configuration
	address := flag.String("server", "", "gRPC server in format host:port")
	flag.Parse()

	//Set up a connection to the server
	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := v1.NewToDoServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t := time.Now().In(time.UTC)
	etc := time.Date(
		2019, 10, 17, 20, 34, 58, 651387237, time.UTC)
	reminder, _ := ptypes.TimestampProto(t)
	estimatedTimeOfCompletion, _ := ptypes.TimestampProto(etc)
	pfx := t.Format(time.RFC3339Nano)

	// actualTimeOfCompletion, _ := ptypes.TimestampProto(t)

	//Create a ToDo Entity
	req1 := v1.CreateRequest{
		Api: apiVersion,
		ToDo: &v1.ToDo{
			Title:                     fmt.Sprintf("title(%s)", pfx),
			Description:               fmt.Sprintf("description (%s)", pfx),
			Status:                    "Started",
			EstimatedTimeOfCompletion: estimatedTimeOfCompletion,
			ActualTimeOfCompletion:    estimatedTimeOfCompletion, //Initial actualTimeOfCompletion is the same as etimatedTimeOfCompletion
			Reminder:                  reminder,
		},
	}
	res1, err := c.Create(ctx, &req1)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}

	log.Printf("Create result: <%+v>\n\n", res1)
	id := res1.Id

	//Read a ToDo entity
	req2 := v1.ReadRequest{
		Api: apiVersion,
		Id:  id,
	}
	res2, err := c.Read(ctx, &req2)
	if err != nil {
		log.Fatalf("Read failed: %v", err)
	}

	log.Printf("Read result: <%+v>\n\n", res2)

	//Update a ToDo entity
	req3 := v1.UpdateRequest{
		Api: apiVersion,
		ToDo: &v1.ToDo{
			Id:                        res2.ToDo.Id,
			Title:                     res2.ToDo.Title,
			Description:               res2.ToDo.Description,
			Status:                    "Completed",
			EstimatedTimeOfCompletion: res2.ToDo.EstimatedTimeOfCompletion,
			Reminder:                  res2.ToDo.Reminder,
		},
	}

	res3, err := c.Update(ctx, &req3)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}
	log.Printf("Update result: <+%v>\n\n", res3)

	//ReadAll ToDo entities
	req4 := v1.ReadAllRequest{
		Api: apiVersion,
	}

	res4, err := c.ReadAll(ctx, &req4)
	if err != nil {
		log.Fatalf("ReadAll failed: %v", err)
	}
	log.Printf("ReadAll result: <%+v>\n\n", res4)

}

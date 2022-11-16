package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	grpcAuction "github.com/magnusblarsen/DSYS_auction_system/proto"
	"google.golang.org/grpc"
)

type Server struct {
	grpcAuction.UnimplementedServicesServer // an interface that the server needs to have

	serverName       string
	port             string
	lamportTimestamp int64
	highestBid       int64
}

type response struct {
	ack        bool
	HighestBid int64
}

var serverName = flag.String("name", "", "Senders name")
var port = flag.String("port", "", "Server port")

func main() {
	flag.Parse()
	configureLog()
	log.Println("::server is starting::")
	launchServer()
}

func configureLog() {
	file, err := os.OpenFile(fmt.Sprintf("server_%s", *serverName), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()

	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)
}

func launchServer() {
	if *serverName == "" || *port == "" {
		//TODO: fejl
	}

	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err)
	}

	grpcServer := grpc.NewServer()
	server := &Server{
		serverName:       *serverName,
		port:             *port,
		highestBid:       0,
		lamportTimestamp: 0,
	}
	grpcAuction.RegisterServicesServer(grpcServer, server)
	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func (s *Server) Bid(ctx context.Context, bidAmount *grpcAuction.BidAmount) (*grpcAuction.Ack, error) {
	if bidAmount.Amount > s.highestBid {
		s.highestBid = bidAmount.Amount
	}

	response := (&grpcAuction.Ack{
		Ack:        true,
		HighestBid: s.highestBid,
	})

	return response, nil
}
func (s *Server) Result(ctx context.Context, resultRequest *grpcAuction.ResultRequest) (*grpcAuction.Outcome, error) {
	return nil, nil
}

// func sendToAllStreams(senderID string, newMessage string) error {
// 	lamportTimestamp++
// 	for _, v := range connectedClientStreams {
// 		err := v.stream.Send(&grpcAuction.ServerMessage{
// 			SenderID:    senderID,
// 			Message:     newMessage,
// 			LamportTime: lamportTimestamp,
// 		})
// 		if err != nil {
// 			return err
// 		}

// 	}
// 	return nil
// }
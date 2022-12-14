package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	grpcAuction "github.com/magnusblarsen/DSYS_auction_system/proto"
	"google.golang.org/grpc"
)

type Server struct {
	grpcAuction.UnimplementedServicesServer // an interface that the server needs to have

	serverName       string
	port             string
	lamportTimestamp int64
	highestBid       int64
	isOver           bool
	highestBidderID  int64
}

var serverName = flag.String("name", "", "Senders name")
var port = flag.String("port", "", "Server port")

func main() {
	flag.Parse()
	configureLog()
	log.Println("::server is starting::")
	server := &Server{
		serverName:       *serverName,
		port:             *port,
		highestBid:       0,
		lamportTimestamp: 0,
		isOver:           true,
	}
	go server.launchServer()
	close := make(chan bool)
	<-close
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

func (server *Server) launchServer() {
	if *serverName == "" || *port == "" {
		log.Fatalln("You must write port and name of the server")
	}

	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", server.port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err)
	}

	grpcServer := grpc.NewServer()
	grpcAuction.RegisterServicesServer(grpcServer, server)
	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}

}

func (s *Server) Bid(ctx context.Context, bidAmount *grpcAuction.BidAmount) (*grpcAuction.Ack, error) {
	log.Println("server bid startet")
	succes := false
	if bidAmount.Amount > s.highestBid && !s.isOver {
		s.highestBid = bidAmount.Amount
		s.highestBidderID = bidAmount.BidderId
		succes = true
	}

	response := (&grpcAuction.Ack{
		Ack: succes,
	})

	return response, nil
}
func (s *Server) Result(ctx context.Context, resultRequest *grpcAuction.ResultRequest) (*grpcAuction.Outcome, error) {

	result := (&grpcAuction.Outcome{
		Outcome: s.highestBid,
		Over:    s.isOver,
		Winner:  s.highestBidderID,
	})

	return result, nil
}

func (s *Server) StartAuction(ctx context.Context, resultRequest *grpcAuction.ResultRequest) (*grpcAuction.Ack, error) {
	success := false

	if s.isOver {
		success = true
		go s.startTimer()
	}

	ack := &grpcAuction.Ack{
		Ack: success,
	}
	return ack, nil
}

func (s *Server) startTimer() {
	s.highestBid = 0
	s.isOver = false
	time.Sleep(15 * time.Second)
	s.isOver = true
}

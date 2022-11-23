package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	grpcChat "github.com/magnusblarsen/DSYS_auction_system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var clientId = flag.Int("id", 0, "client id")

type Client struct {
	serverConns map[int64]grpcChat.ServicesClient
	clientId    int
}

func main() {
	flag.Parse()

	file, err := os.OpenFile(fmt.Sprintf("logs\\client_%d", *clientId), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()

	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	clientHandle := &Client{
		serverConns: make(map[int64]grpcChat.ServicesClient),
		clientId:    *clientId,
	}

	fmt.Println("--- CLIENT APP ---")
	clientHandle.connectToServers()

	clientHandle.parseInput()
	//FIXME: Lige nu har vi ikke nogen defer conn.Close()
	// for _, s := range clientHandle.serverConns {
	// 	defer s.Close()
	// }
}

func (c *Client) connectToServers() {
	var opts []grpc.DialOption
	opts = append(
		opts, grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	portfile, err := os.Open("client/server-addresses.txt")
	if err != nil {
		log.Printf("could not read text file with servers: %v", err)
	}
	defer portfile.Close()
	fileScanner := bufio.NewScanner(portfile)
	for fileScanner.Scan() {
		port, convErr := strconv.ParseInt(fileScanner.Text(), 10, 64)
		if convErr != nil {
			log.Printf("Could not convert: %v", convErr)
		}

		log.Printf("id %v is trying to dial: %v\n", *clientId, port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), opts...)
		if err != nil {
			log.Fatalf("Could not connect: %v", err)
		}
		//defer conn.Close()
		clientApi := grpcChat.NewServicesClient(conn)

		c.serverConns[port] = clientApi
		log.Printf("Connection successful for port %v", port)

	}
}

func (c *Client) parseInput() {
	reader := bufio.NewReader(os.Stdin)
	log.Println("Type your bid. example: bid 200")
	fmt.Println("--------------------")

	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input

		inputvalues := strings.Split(input, " ")
		switch strings.ToLower(inputvalues[0]) {
		case "bid":
			//Convert string to int64, return error if the int is larger than 64bit or not a number
			val, err := strconv.ParseInt(inputvalues[1], 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			c.Bid(val, int64(*clientId))
		case "result":
			c.Result()
		case "start":
			c.StartAuction()
		default:
			log.Println("type a valid input")
		}
	}
}

func (c *Client) Bid(val int64, bidderID int64) {
	bid := &grpcChat.BidAmount{
		Amount:   val,
		BidderId: bidderID,
	}

	var someAck *grpcChat.Ack
	writeSuccesses := 0
	for _, v := range c.serverConns {
		serverConn := v
		ack, err := serverConn.Bid(context.Background(), bid)
		if err != nil {
			log.Println("One server has crashed. Skipping")
			fmt.Print("-> ")
		} else {
			writeSuccesses++
			someAck = ack
		}
	}
	if writeSuccesses < len(c.serverConns)-1 {
		log.Fatal("Too many server-crashes to handle arhhhhh")
	}
	if someAck.Ack {
		log.Printf("%v placed a %v kr bid\n", bidderID, val)
	} else {
		log.Printf("Invalid Bid: Either the auction is over or you bid below the current leader\n")
	}

}

func (c *Client) Result() {
	request := &grpcChat.ResultRequest{}
	//outcomeChan := make(chan *grpcChat.Outcome)

	var someOutcome *grpcChat.Outcome
	readSuccesses := 0
	for _, v := range c.serverConns {
		serverConn := v
		outcome, err := serverConn.Result(context.Background(), request)
		if err != nil {
			log.Println("One server has crashed. Skipping")
			fmt.Print("-> ")
			continue
		}
		readSuccesses++
		someOutcome = outcome
		if readSuccesses > 1 {
			break
		}
	}
	if readSuccesses < 2 {
		log.Fatal("Too many server-crashes to handle arhhhhh")
	}
	if someOutcome.Over {
		log.Printf("The auction is over and the higest bid was %v kr by bidder %v\n", someOutcome.Outcome, someOutcome.Winner)
	} else {
		log.Printf("The current highest bid is %v kr by bidder %v\n", someOutcome.Outcome, someOutcome.Winner)
	}
}

func (c *Client) StartAuction() {
	request := &grpcChat.ResultRequest{}
	var someAck *grpcChat.Ack
	writeSuccesses := 0
	for _, v := range c.serverConns {
		serverConn := v
		outcome, err := serverConn.StartAuction(context.Background(), request)
		if err != nil {
			log.Println("One server has crashed. Skipping")
			fmt.Print("-> ")
			continue
		} else {
			writeSuccesses++
			someAck = outcome
		}
	}
	if writeSuccesses < len(c.serverConns)-1 {
		log.Fatal("Too many server-crashes to handle arhhhhh")
	}
	if !someAck.Ack {
		log.Println("Auction is already running")
	} else {
		log.Println("the Auction has started")
	}
}

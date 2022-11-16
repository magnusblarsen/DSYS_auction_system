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

var clientsName = flag.String("name", "default", "Senders name")
var clientId = flag.String("id", "", "client id")

var client grpcChat.ServicesClient
var ServerConns map[int32]grpcChat.ServicesClient

func main() {
	flag.Parse()

	file, err := os.OpenFile(fmt.Sprintf("client_%s", *clientsName), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()

	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	fmt.Println("--- CLIENT APP ---")
	go connectToServers()

	parseInput()
}

func connectToServers() {
	ServerConns = make(map[int32]grpcChat.ServicesClient)
	var opts []grpc.DialOption
	opts = append(
		opts, grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	portfile, err := os.Open("server-addresses.txt")
	defer portfile.Close()
	if err != nil {
		log.Printf("could not read text file with servers: %v", err)
	}
	fileScanner := bufio.NewScanner(portfile)
	for fileScanner.Scan() {
		port, _ := strconv.ParseInt(fileScanner.Text(), 10, 32)
		port32 := int32(port)

		var conn *grpc.ClientConn
		log.Printf("id %v is trying to dial: %v\n", clientId, port)
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%v", port), opts...)
		if err != nil {
			log.Fatalf("Could not connect: %v", err)
		}
		defer conn.Close()
		client = grpcChat.NewServicesClient(conn)
		ServerConns[port32] = client
	}
}

func parseInput() {
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
			clientId, err := strconv.ParseInt(*clientsName, 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			Bid(val, clientId)
		case "result":
			Result()
		default:
			log.Println("type a valid input")
		}
	}
}

func Bid(val int64, bidderID int64) {
	bid := &grpcChat.BidAmount{
		Amount:   val,
		BidderId: bidderID,
	}

	ack, _ := client.Bid(context.Background(), bid)
	if ack.Ack {
		log.Printf("%v placed a %v kr bid\n", bidderID, val)
	} else {
		log.Printf("Your bid has to be higher than the current leader %v kr\n", ack.HighestBid)
	}
}

func Result() {
	request := &grpcChat.ResultRequest{}
	outcome, _ := client.Result(context.Background(), request)
	if outcome.Over {
		log.Printf("The auction is over and the higest bid was %v kr\n", outcome.Outcome)
	} else {
		log.Printf("The current highest bid is %v kr\n", outcome.Outcome)
	}
}

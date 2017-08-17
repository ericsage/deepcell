package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ericsage/deepcell/dc"
	"google.golang.org/grpc"
)

type Response struct {
	Data   *dc.Reply `json:"data"`
	Errors []string  `json:"errors"`
}

var (
	tls        = getenv("tls", "false")
	serverAddr = getenv("SERVER_ADDRESS", "biancone.ucsd.edu")
	serverPort = getenv("SERVER_PORT", "5000")
	labels     = []string{"ontology"}
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func knockout(growth bool, ontology string, genes []string) *dc.Reply {
	log.Println("About to open connection")
	address := serverAddr + ":" + serverPort
	log.Println(address)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		panic("Could not connect to the deep cell server!")
	}
	log.Println("Dialed server")
	defer conn.Close()
	client := dc.NewDeepCellClient(conn)
	req := &dc.Request{
		Genes:    genes,
		Ontology: ontology,
		Growth:   growth,
	}
	log.Println("About to process request")
	rep, err := client.Run(context.Background(), req)
	if err != nil {
		panic(err)
	}
	return rep
}

func deepcellHandler(w http.ResponseWriter, r *http.Request) {
	defer catchError()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic("Could not read request body!")
	}

	url := r.URL
	ontology := url.Query().Get("ontology")
	if ontology == "" {
		ontology = "GO"
	}
	growth := url.Query().Get("growth")
	var genes []string
	err = json.Unmarshal(body, &genes)
	if err != nil {
		panic("Could not decode json into genes list!")
	}
	termReply := knockout(ontology, genes, growth == "true")
	res := &Response{
		Data:   termReply,
		Errors: []string{},
	}
	json.NewEncoder(w).Encode(res)
}

func catchError() {
	if r := recover(); r != nil {
		//Send response json error response here
		log.Println(r)

	}
}

func main() {
	log.Println("Starting server...")
	http.HandleFunc("/", deepcellHandler)
	http.ListenAndServe(":80", nil)
}

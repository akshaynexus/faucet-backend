package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/akshaynexus/go-bitcoind"
	"github.com/albrow/forms"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
)

const (
	//ServerHost is the host
	ServerHost = "127.0.0.1"
	//ServerPort is the port
	ServerPort = 51925
	//USER is the USER
	USER = "user"
	//PASSWD is the password
	PASSWD = "pass"
	//USESSL is the setter to use ssl or not
	USESSL = false
)

//StatsStruct to send on stats req
type StatsStruct struct {
	Balance   float64 `json:"balance"`
	Blocks    uint64  `json:"blocks"`
	Totalsent int     `json:"totalsent"`
}

//SendReturn to send on send action
type SendReturn struct {
	Error      string `json:"error"`
	TxID       string `json:"txid"`
	SentAmount int    `json:"sentamt"`
}

//SendReq to send on send action
type SendReq struct {
	Address string `form:"address" binding:"required"`
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage endpoint hit")
}

func handleError(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func handleRequests() {
	endpoints := map[string]http.HandlerFunc{
		"/": prepareStats,
		"/send": sendMoney,
	}
	port := ":3333"
	var router = mux.NewRouter()
    router.Use(commonMiddleware)
	for endpoint, f := range endpoints {
		router.HandleFunc(endpoint, f)
	}

    headersOk := handlers.AllowedHeaders([]string{"Authorization"})
    originsOk := handlers.AllowedOrigins([]string{"*"})
    methodsOk := handlers.AllowedMethods([]string{"GET", "POST"})

    fmt.Printf("Server is running at http://localhost%s\n", port)
    log.Fatal(http.ListenAndServe(port, handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}

func sendMoney(w http.ResponseWriter, r *http.Request) {
	bc, err := bitcoind.New(ServerHost, ServerPort, USER, PASSWD, USESSL)
	handleError(err)
	userData, err := forms.Parse(r)
	addr := userData.Get("address") // addr will be "" if parameter is not set
	if addr == "" {
		dataReturn := &SendReturn{Error: "No address found", TxID: "", SentAmount: 0}
		//Convert to json and send
		json.NewEncoder(w).Encode(dataReturn)
	} else {
		log.Printf("Addr is: " + addr)
		vresponse, err := bc.ValidateAddress(addr)
		handleError(err)
		vamountosend := 50
		//Prepare stats
		if vresponse.IsValid {
			log.Printf("sending " + strconv.Itoa(vamountosend))
			txid, err := bc.SendToAddress(addr, float64(vamountosend), "faucet payout", addr)
			handleError(err)
			dataReturn := &SendReturn{Error: "", TxID: txid, SentAmount: vamountosend}
			//Convert to json and send
			json.NewEncoder(w).Encode(dataReturn)

		} else {
			dataReturn := &SendReturn{Error: "Invalid address given", TxID: "", SentAmount: 0}
			//Convert to json and send
			json.NewEncoder(w).Encode(dataReturn)
		}
	}

}

func prepareStats(w http.ResponseWriter, r *http.Request) {
	bc, err := bitcoind.New(ServerHost, ServerPort, USER, PASSWD, USESSL)
	handleError(err)
	//Prepare stats
	balance, err := bc.GetBalance("*", 0)
	handleError(err)
	blocks, err := bc.GetBlockCount()
	handleError(err)
	stats := &StatsStruct{Balance: balance, Blocks: blocks, Totalsent: 100}
	//Convert to json and send
	json.NewEncoder(w).Encode(stats)
}

func main() {
	handleRequests()
}

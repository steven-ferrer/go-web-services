package main

import (
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

//RPCCAPIArguments arguments as usual
type RPCCAPIArguments struct {
	Message string
}

//RPCAPIResponse response as usual
type RPCAPIResponse struct {
	Message string
}

//StringService service for strings
type StringService struct{}

//Length length service
func (h *StringService) Length(r *http.Request, arguments *RPCCAPIArguments, reply *RPCAPIResponse) error {
	reply.Message = fmt.Sprintf("Your string is "+
		"%d chars long", utf8.RuneCountInString(arguments.Message)) + " chars long"
	return nil
}

func main() {
	fmt.Println("Starting service...")
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(new(StringService), "")
	http.Handle("/rpc", s)
	http.ListenAndServe(":10000", nil)
}

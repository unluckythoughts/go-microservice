package main

import (
	"fmt"

	"github.com/unluckythoughts/go-microservice/v2/integrations/pan"
	"github.com/unluckythoughts/go-microservice/v2/tools/context"
	"github.com/unluckythoughts/go-microservice/v2/tools/logger"
)

func main() {
	l := logger.New(logger.Options{})
	s := pan.New(pan.Options{
		Code:  "<your-imwallet-code>",
		Token: "<your-imwallet-token>",
	})

	_, err := s.IMWalletGetPANDetails(context.NewContext(l), "APGPG5332L")
	if err != nil {
		fmt.Println("Error fetching PAN details:", err)
		return
	}
}

package configs

import (
	"fmt"
	"os"
	"strconv"
)

func GetEnvironment() string {
	return os.Getenv("ENV")
}

func GetServerPort() string {
	res, err := strconv.Atoi(os.Getenv("SERVER_PORT"))

	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%d", res)
}

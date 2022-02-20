package videoinsight_test

import (
	"log"
	"os"

	"github.com/korylprince/go-videoinsight"
)

func Example() {
	client := videoinsight.NewClient("http", "example.com", 9000)
	if err := client.Authenticate("username", "password", 600); err != nil {
		log.Fatalln("could not authenticate:", err)
	}
	cameras, err := client.Cameras()
	if err != nil {
		log.Fatalln("could not get cameras:", err)
	}

	snapshot, err := client.Snapshot(cameras[0].ID)
	if err != nil {
		log.Fatalln("could not get snapshot:", err)
	}

	if err = os.WriteFile("snapshot.jpg", snapshot, 0644); err != nil {
		log.Fatalln("could not write snapshot:", err)
	}
}

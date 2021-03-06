package sproducer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/4lie/nats101/model"
	"github.com/nats-io/stan.go"
	"github.com/spf13/cobra"
)

//nolint:funlen
func main(server string, cid string) {
	rand.Seed(time.Now().UnixNano())

	// nolint:gosec
	id := rand.Int63()

	nc, err := stan.Connect(cid, fmt.Sprintf("elahe-%d", id), stan.NatsURL(server))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Connected to %s from %v\n", nc.NatsConn().ConnectedAddr(), nc.NatsConn().DiscoveredServers())

	defer nc.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)

			continue
		}

		line = strings.TrimSuffix(line, "\n")

		splited := strings.SplitN(line, " ", 2)

		var cmd, args string

		if len(splited) > 1 {
			cmd, args = splited[0], splited[1]
		} else {
			cmd = splited[0]
		}

		switch cmd {
		case "send":
			splited := strings.SplitN(args, " ", 2)
			from, message := splited[0], splited[1]

			b, err := json.Marshal(model.Message{
				From:      from,
				Text:      message,
				CreatedAt: time.Now(),
			})
			if err != nil {
				log.Fatal(err)
			}

			if err := nc.Publish("message", b); err != nil {
				log.Fatal(err)
			}
		case "exit":
			return
		default:
			fmt.Println("Please enter valid command")
		}
	}
}

// Register producer command.
func Register(root *cobra.Command, server *string) {
	cmd := &cobra.Command{
		Use:   "sproducer",
		Short: "Produce messages to streaming NATS",
		Run: func(cmd *cobra.Command, args []string) {
			cid, err := cmd.Flags().GetString("cluster")
			if err != nil {
				log.Printf("invalid cluster argument %s", err)
			}

			main(*server, cid)
		},
	}

	cmd.Flags().StringP("cluster", "c", "4lie", "nats streaming cluster-id")

	root.AddCommand(cmd)
}

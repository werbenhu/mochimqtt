package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/logrusorgru/aurora"

	mqtt "github.com/mochi-co/mqtt/server"
	"github.com/mochi-co/mqtt/server/listeners"
	"github.com/mochi-co/mqtt/server/persistence/bolt"
)

func main() {

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	fmt.Println(aurora.Magenta("Mochi MQTT Server initializing..."), aurora.Cyan("Persistence"))

	server := mqtt.New()
	tcp := listeners.NewTCP("t1", ":1883")
	err := server.AddListener(tcp, &listeners.Config{
		Auth: new(Auth),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Add the persistent store to the server.
	err = server.AddStore(bolt.New("mochi-test.db", nil))
	if err != nil {
		log.Fatal(err)
	}

	// Start broker...
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()
	fmt.Println(aurora.BgMagenta("  Started!  "))

	// Wait for signals...
	<-done
	fmt.Println(aurora.BgRed("  Caught Signal  "))

	// End gracefully.
	server.Close()
	fmt.Println(aurora.BgGreen("  Finished  "))

}

// Auth is an example auth provider for the server.
type Auth struct{}

// Auth returns true if a username and password are acceptable.
// Auth always returns true.
func (a *Auth) Authenticate(user, password []byte) bool {
	return true
}

// ACL returns true if a user has access permissions to read or write on a topic.
// ACL is used to deny access to a specific topic to satisfy Test.test_subscribe_failure.
func (a *Auth) ACL(user []byte, topic string, write bool) bool {
	if topic == "test/nosubscribe" {
		return false
	}
	return true
}

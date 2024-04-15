package gpssocket

import (
	"fmt"
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
)

// UserConnection represents a user's connection information
type UserConnection struct {
	SocketID string
}

var (
	server *socketio.Server
)

// SocketInit initializes the Socket.IO server
func SocketInit() {
	server = socketio.NewServer(nil)

	// Handle connections
	server.OnConnect("/", func(s socketio.Conn) error {
		log.Println("New connection:", s.ID())
		return nil
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		server.LeaveAllRooms("/", s)
		log.Println("Disconnected:", s.ID(), "Reason:", reason)
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnEvent("/", "joinRoute", func(s socketio.Conn, routeID string) {
		server.LeaveAllRooms("/", s)
		server.JoinRoom("/", routeID, s)
		log.Println("server rooms", server.Rooms("/"))

		fmt.Printf("Bus joined route %s\n", routeID)
	})

	http.Handle("/socket.io/", server)
	//http.Handle("/", http.FileServer(http.Dir(".")))

	// Start the Socket.IO server
	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer server.Close()

	log.Println("Serving at localhost:8090...")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

// SendMessageToUser sends a message to a specific user
func SendMessageToUser(roomName string, eventName string, data string) {
	// Send message to user's connection
	success := server.BroadcastToRoom("/", roomName, eventName, data)
	if !success {
		log.Println("Failed to broadcast data to room:", roomName)

	} else {
		log.Println("Broadcast successful to room:", roomName)
	}
}

package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	fmt.Println("connected.port@ 6379")

	//setup TCP: Transmission Control Protocol server. This server reads in RESP data from
	//redis-cli. The listening port is 6379. On receiving and accepting incoming
	//connection request from redis cli, establish a communication channel with redis-cli
	tsrv, err := net.Listen("tcp", ":6379")
	//check if error occured during server setup
	if err != nil {
		fmt.Println(err)
		return
	}

	aof, err := NewAof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aof.Close()

	// Performing operations from the AOF file before executing them in memory offers
	// data durability, replayability, and consistency in database systems. By logging
	// every operation to disk first, potential data loss due to system crashes or restarts
	// is mitigated. Replaying operations from the AOF file during system recovery ensures
	// that the database state is accurately reconstructed. Additionally, executing operations
	// from the AOF file guarantees that the in-memory database reflects all logged operations,
	// maintaining data consistency. Asynchronous execution of AOF file operations can improve
	// system performance by separating disk I/O from other application tasks. Furthermore,
	// inspecting the AOF file allows for debugging and monitoring of database activity, providing
	// insights into the history of operations. In summary, leveraging the AOF file for operations
	// before executing them in memory enhances data durability, consistency, and system
	/// performance in database management
	aof.Read(func(value Value) {
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		handler(args)
	})
	//Accepts incoming connections ('aconn') from clients on TCP listener ('tsrv').
	aconn, err := tsrv.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	//defer connection closing before function exits
	defer aconn.Close()

	for {
		// create new instance of a pointer to
		// an RESP struct with aconn
		redis_msg := newrESP(aconn)
		// read RESP struct for redis_msg using Read
		value, err := redis_msg.Read()
		if err != nil {
			fmt.Println(err)
			return
		}
		// Ensure the message is of type array
		if value.typ != "array" {
			// print error if not array and
			// continue to next iteration
			fmt.Println("Invalid request, expected array")
			continue
		}
		// Ensure message is not empty
		if len(value.array) == 0 {
			// print error if empty
			// and continue to the next iteration
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}

		// This line of code converts the first element of an array,
		// accessed via `value.array[0].bulk`, to uppercase using the
		// `strings.ToUpper()` function. The resulting uppercase string
		// is assigned to the variable `command`.
		command := strings.ToUpper(value.array[0].bulk)
		// set array[1:] to args
		args := value.array[1:]
		// create  a new instance
		writer := NewWriter(aconn)
		// check handler validity
		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(Value{typ: "string", str: ""})
			continue
		}
		if command == "SET" || command == "HSET" {
			aof.Write(value)
		}
		// return results on arguments
		result := handler(args)
		writer.Write(result)
	}
}

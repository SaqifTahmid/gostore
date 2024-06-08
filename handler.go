package main

import (
	"sync"
)

// The Handlers map is a core part of the command processing mechanism
// for GO server. It maps command names (like "PING", "SET", "GET")
// to their corresponding handler functions.
var Handlers = map[string]func([]Value) Value{
	// "PING": Returns a "PONG" response
	"PING": ping,
	// "SET": Stores a key-value pair
	"SET": set,
	// "GET": Retrieves the value for a given key
	"GET": get,
	// "HSET": Sets a field in a hash stored at a key
	"HSET": hset,
	// "HGET": Retrieves a field from a hash stored at a key
	"HGET": hget,
	// "HGETALL": Retrieves all fields and values of a hash stored at a key
	"HGETALL": hgetall,
}

// ping function takes a slice of Value structs as arguments and returns a Value struct.
// The function is designed to handle the PING command in Redis.
func ping(args []Value) Value {
	if len(args) == 0 {
		// If there are no arguments, return a Value with type "string" and the content "PONG"
		return Value{typ: "string", str: "PONG"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

// SETs is a global map variable that stores key-value pairs.
// It is intended to hold string keys and string values.
var SETs = map[string]string{}

// SETsMu is a global read-write mutex variable used for synchronization.
// It provides exclusive access to the SETs map to prevent race conditions
// when reading from or writing to the map concurrently from multiple goroutines.
var SETsMu = sync.RWMutex{}

// set func echoes the SET function from a redis database
func set(args []Value) Value {
	// check for arguments error
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	}
	// key from command
	key := args[0].bulk
	// val from command
	value := args[1].bulk
	// Acquires an exclusive lock (Lock()) on the mutex SETsMu, ensuring mutual exclusion.
	// This prevents other goroutines from accessing or modifying the map concurrently
	SETsMu.Lock()
	SETs[key] = value
	// Releases the lock (Unlock()) on the mutex SETsMu after the update operation is
	// completed. Releasing the lock allows other goroutines to acquire it and perform
	// their operations on the map
	SETsMu.Unlock()
	// If the key exists, return OK
	return Value{typ: "string", str: "OK"}
}

// get function simulates the GET command from a Redis-like database.
// It retrieves the value associated with the specified key from the database.
// If the key does not exist, it returns a null value.
func get(args []Value) Value {
	// Check for the correct number of arguments
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	}

	// Extract the key from the command arguments
	key := args[0].bulk

	// Acquire a read lock (RLock()) on the mutex SETsMu to allow concurrent reads
	SETsMu.RLock()
	// Retrieve the value associated with the key from the map SETs
	value, ok := SETs[key]
	// Release the read lock (RUnlock()) on the mutex SETsMu after the read operation
	SETsMu.RUnlock()

	// If the key does not exist in the map, return a null value
	if !ok {
		return Value{typ: "null"}
	}

	// If the key exists, return the value associated with it
	return Value{typ: "bulk", bulk: value}
}

// HSETs is a map representing a Redis-like hash set data structure.
var HSETs = map[string]map[string]string{}

// HSETsMu is a read-write mutex used for synchronization when accessing the HSETs map.
// It provides exclusive access to the map to prevent race conditions when reading from
// or writing to the map concurrently from multiple goroutines.
var HSETsMu = sync.RWMutex{}

// The HSET command is used to set the value of a field within a hash stored at a specific key.
// It operates on Redis hash data structures, which allow for the storage of multiple field-value pairs under a single key.
func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}
	}
	// access hash table
	hash := args[0].bulk
	key := args[1].bulk
	value := args[2].bulk

	// Acquire an exclusive lock (Lock()) on the mutex HSETsMu to ensure mutual exclusion
	HSETsMu.Lock()
	if _, ok := HSETs[hash]; !ok {
		HSETs[hash] = map[string]string{}
	}
	HSETs[hash][key] = value
	// Release the lock (Unlock()) on the mutex HSETsMu after the update operation
	HSETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

// hget is a function that retrieves the value associated with a specified key from
// a hash in the in-memory database.It takes an array of arguments, where the first
// argument is the hash name and the second argument is the key. If the number of
// arguments is not equal to 2, it returns an error message indicating the incorrect
// number of arguments.It then retrieves the value corresponding to the provided key
// from the specified hash.

// If the key does not exhouist in the hash, it returns a null value.
// Otherwise, it returns the value associated with the key as a bulk response.

func hget(args []Value) Value {
	// Check if the number of arguments is not equal to 2
	if len(args) != 2 {
		// Return an error message indicating the incorrect number of arguments
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hget' command"}
	}

	// Extract the hash name and key from the arguments
	hash := args[0].bulk
	key := args[1].bulk

	// Read lock to access the hash map storing the hash sets
	HSETsMu.RLock()
	// Retrieve the value associated with the key from the hash set
	value, ok := HSETs[hash][key]
	// Release the read lock
	HSETsMu.RUnlock()

	// Check if the key exists in the hash set
	if !ok {
		// If the key does not exist, return a null value
		return Value{typ: "null"}
	}

	// If the key exists, return the associated value
	return Value{typ: "bulk", bulk: value}
}

// hgetall is a function that retrieves all key-value pairs from a hash in the in-memory database.
// It takes an array of arguments, where the only argument is the hash name.
// If the number of arguments is not equal to 1, it returns an error message indicating the incorrect number of arguments.
// It then retrieves all key-value pairs from the specified hash.
// If the hash does not exist, it returns a null value.
// Otherwise, it returns an array containing all key-value pairs as bulk responses, alternating between keys and values.
func hgetall(args []Value) Value {
	// Check if the number of arguments is not equal to 1
	if len(args) != 1 {
		// Return an error message indicating the incorrect number of arguments
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hgetall' command"}
	}

	// Extract the hash name from the arguments
	hash := args[0].bulk

	// Read lock to access the hash map storing the hash sets
	HSETsMu.RLock()
	// Retrieve the hash set associated with the hash name
	value, ok := HSETs[hash]
	// Release the read lock
	HSETsMu.RUnlock()

	// Check if the hash exists
	if !ok {
		// If the hash does not exist, return a null value
		return Value{typ: "null"}
	}

	// Initialize an empty array to store key-value pairs
	values := []Value{}
	// Iterate over all key-value pairs in the hash set
	for k, v := range value {
		// Append the key and value as bulk responses to the values array
		values = append(values, Value{typ: "bulk", bulk: k})
		values = append(values, Value{typ: "bulk", bulk: v})
	}

	// Return an array containing all key-value pairs
	return Value{typ: "array", array: values}
}

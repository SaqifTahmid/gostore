// // File for derisilization of message received from redis-cli.
package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// define constants for Redis Serialization Protocol
const (
	//STRING ('+'): This represents a simple string response.
	//It's used for simple messages like "+OK\r\n".
	STRING = '+'
	//ERROR ('-'): This represents an error message.
	//It's used to indicate that something went wrong, like "-Error message\r\n".
	ERROR = '-'
	//INTEGER (':'): This represents an integer. It's used to return numeric values,
	//like ":1000\r\n"
	INTEGER = ':'
	//BULK ('$'): This represents a bulk string. It's used for strings that might
	//include spaces or special characters, like "$6\r\nfoobar\r\n".
	BULK = '$'
	//ARRAY ('*'): This represents an array. It's used to return a list of elements,
	//like "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n".
	ARRAY = '*'
)

// define struct for Values for parsing and represing Redis protocol in GO
type Value struct {
	//data type for value
	typ string
	//value of string from simple strings
	str string
	//value of integer received from integers
	num int
	//store strings from bulk strings
	bulk string
	//holds values from arrays
	array []Value
}

// struct for pointer to memory to avoid copies
type rESP struct {
	// reader serves as a memory pointer for bufio.Reader
	// bufio.reader is a wrapper for io.reader to buffer
	// incoming byte slice stream in-memory
	reader *bufio.Reader
}

// newrESP receives data in io.Reader as data stream received from redis-cli
// returns a pointer to it to avoid copies. This serves for cost
// reduction purposes, easier modification and code readability
func newrESP(rd io.Reader) *rESP {
	// bufio.NewReader is a function provided by Go's bufio package. It is used to
	// create a new bufio.Reader object that wraps an existing io.Reader, providing
	// buffering and additional functionality for reading data from the input sourc
	return &rESP{reader: bufio.NewReader(rd)}
}

// func for readLine method which is bound to an instance of rESP struct
// it returns line as a list of byte, number of bytes read and error if it occurs
func (r *rESP) readLine() (line []byte, n int, err error) {
	// start infinite loop to read bytes one by one
	for {
		// read single byte from reader
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		// increment count by one for every byte read
		n += 1
		// append byte to list called line
		line = append(line, b)
		// termination condition
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}

	// return list of byte
	return line[:len(line)-2], n, nil
}

// func for readInteger method which is bound to an instance of rESP struct
// it returns integer, number of bytes read and error if it occurs
func (r *rESP) readInteger() (x int, n int, err error) {

	// read line from redis-cli message stored in-memory
	line, n, err := r.readLine()
	if err != nil {

		// if error reading line
		return 0, 0, err
	}

	// parse/interpret byte slice as string and convert to integer.
	// ignore a
	i64, err := strconv.ParseInt(string(line), 10, 64)

	// return error for parse issue
	if err != nil {
		return 0, n, err
	}
	// return parsed int64 as int type
	return int(i64), n, nil
}

// func bound to pointer fr RESP value from
// input stream recevied from redis cli
// It returns a Value error for not Value Type or
// access readarray/readbulk
func (r *rESP) Read() (Value, error) {
	// read single byte from input stream
	_type, err := r.reader.ReadByte()
	//if error return empty Value
	if err != nil {
		return Value{}, err
	}
	// determine type of value based on byte read
	switch _type {
	//check if byte is array
	case ARRAY:
		return r.readArray()
	//check if byte is bulk
	case BULK:
		return r.readBulk()
	//byte is neither
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// func to read array  from input stream recevied
// from redis-cli it is bound to a pointer to rESP struct
func (r *rESP) readArray() (Value, error) {
	// set v to Value struct
	v := Value{}
	v.typ = "array"
	// read length of array using readInteger method
	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}
	// for each line, parse and read the value
	v.array = make([]Value, 0)
	// loop continues till array length reached
	for i := 0; i < len; i++ {
		//call Read on every line in array
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		// append parsed value to array
		v.array = append(v.array, val)
	}
	return v, nil
}

// func to  read length of bulk string
func (r *rESP) readBulk() (Value, error) {
	// start an instance of Value
	v := Value{}
	// v.type is set to bulk
	v.typ = "bulk"
	// check for length
	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}
	// create  byte slice to hold bulk string
	bulk := make([]byte, len)
	// parse bulk
	r.reader.Read(bulk)
	v.bulk = string(bulk)
	// Read the trailing CRLF
	r.readLine()
	//return the value
	return v, nil
}

//Marshal Value to Byte to trasmit over network.
//respond to the client with RESP and write the Writer.

// func resposible for calling appropriate method
// to convert value to byte
func (v Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshallNull()
	case "error":
		return v.marshallError()
	default:
		return []byte{}
	}
}

// func to marshalString for simple string
// for the Value type
func (v Value) marshalString() []byte {
	// declare a byte slice called bytes
	var bytes []byte
	// Appends the STRING identifier to the bytes slice.
	// In the RESP protocol, a simple string is prefixed with a +
	// character (assuming STRING is a constant representing this)
	bytes = append(bytes, STRING)
	// Appends the actual string content stored in the str field of
	// the Value struct to the bytes slice. The ... is a variadic
	// argument syntax that spreads the string into individual bytes.
	bytes = append(bytes, v.str...)
	// Appends a carriage return (\r) and line feed (\n) to the bytes
	// slice. This marks the end of the RESP string.
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	//Appends the BULK identifier to the bytes slice.
	//In the RESP protocol, a bulk string is prefixed with a $
	//character (assuming BULK is a constant representing this)
	bytes = append(bytes, BULK)
	//this is appending the length of the bulk string v.bulk
	//as individual bytes to the bytes slice
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	//Appends the actual bulk string content stored in the bulk field of
	//the Value struct to the bytes slice.
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	///store length of array
	len := len(v.array)
	var bytes []byte
	//Appends the ARRAY identifier to the bytes slice.
	//In the RESP protocol, aan array is prefixed with a *
	//character (assuming ARRAY is a constant representing this)
	bytes = append(bytes, ARRAY)
	//this is appending the length of the array len
	//as individual bytes to the bytes slice
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')
	//use a loop to add element at i and pass it through marshal
	for i := 0; i < len; i++ {
		bytes = append(bytes, v.array[i].Marshal()...)
	}

	return bytes
}

// marshallError converts the Value representing an error message
// to its RESP (Redis Serialization Protocol) representation as a byte slice.
// It prefixes the error message with the ERROR identifier and terminates
// it with the CRLF (Carriage Return + Line Feed) sequence.
func (v Value) marshallError() []byte {
	// Initialize a byte slice to store the RESP representation
	var bytes []byte
	// Append the ERROR identifier to the byte slice.
	// In the RESP protocol, an error is prefixed with a '-' character
	// (assuming ERROR is a constant representing this)
	bytes = append(bytes, ERROR)
	// Append the error message string to the byte slice
	bytes = append(bytes, v.str...)
	// Append the CRLF (Carriage Return + Line Feed) sequence to indicate
	// the end of the error message
	bytes = append(bytes, '\r', '\n')

	// Return the byte slice representing the RESP-encoded error message
	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}

// Writer properties
// create a wrtier struct to take io.writer
type Writer struct {
	writer io.Writer
}

// create a newinstance of the writer struct and
// return a pointer to the struct
func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

// func binds Write method to a pointer type for a Writer struct
// and returns error if there is an error
func (w *Writer) Write(v Value) error {
	// Marshal the Value v into its RESP representation as a byte slice
	var bytes = v.Marshal()
	// Write the byte slice to the underlying io.Writer
	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

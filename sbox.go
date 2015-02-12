        
package main

import (
    "os"
    "fmt"
    "strconv"
    "github.com/goburrow/modbus"
)

const (
    quantity = 9
    on = uint16(0xFF00)
    off = uint16(0x0000)
)

var (
    host string
    action string
)

func main() {
    // validate command
    if len(os.Args) == 1 || len(os.Args) == 2 && os.Args[1] == "help" {
        fmt.Println(`
    Usage:
        sbox [host] [action] [address]
    Action:
        read-coil           READ_COILS
        write-coil          WRITE_SINGLE_COILS
        read-input          READ_DESCRETE_INPUTS
    Sample:
        sbox /dev/ttyUSB0 read-coil 500
        sbox /dev/ttyUSB0 read-input 410
        sbox /dev/ttyUSB0 write-coil-on 500
        sbox /dev/ttyUSB0 write-coil-off 500
        `)

        return
    } else if(len(os.Args) < 3) {
        fmt.Println("ERROR: host & action not specified")

        return
    }

    // set host
    host = os.Args[1]
    action = os.Args[2]

    // check if type of client
    // its tcp by default
    client := modbus.TCPClient(host)
    if string(host[0]) == "/" {
        client = modbus.ASCIIClient(host)        
    }

    // check action param
    if len(os.Args) == 3 || os.Args[3] == "" {
        fmt.Println("ERROR: address expected")

        return
    }

    // assign address
    address64, err := strconv.ParseUint("0x" + os.Args[3], 0, 16)
    if err != nil {
        fmt.Println("ERROR: invalid address")

        return
    }
    
    // castback to 16
    address := uint16(address64)

    // check action argument
    switch action {
    case "read-coil":
        result, err := client.ReadCoils(address, quantity)
        if err != nil || result == nil {
            fmt.Println("ERROR:", err, result)

            return
        }

        fmt.Println(result)

        return
    case "read-input":
        result, err := client.ReadDiscreteInputs(address, quantity)
        if err != nil || result == nil {
            fmt.Println("ERROR:", err, result)

            return
        }

        fmt.Println(result)

        return
    case "write-coil-on":
        result, err := client.WriteSingleCoil(address, on)
         if err != nil || result == nil {
            fmt.Println("ERROR:", err, result)

            return
        }

        fmt.Println(result)

        return
    case "write-coil-off":
        result, err := client.WriteSingleCoil(address, off)
         if err != nil || result == nil {
            fmt.Println("ERROR:", err, result)

            return
        }

        fmt.Println(result)

        return
    default:
        fmt.Println("ERROR: undefined action for", action)

        return
    }
}
package main

import (
    "fmt"
    "time"
    "encoding/csv"
    "os"
    "os/exec"
)

const (
    device = "/dev/ttyUSB0"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Please define csv file!")
        exit()
    }

    csv := os.Args[1]

    parseCsv(csv)
}

func parseCsv(path string) {
    file, err := os.Open(path)
    if err != nil {
        // err is printable
        // elements passed are separated by space automatically
        fmt.Println("Error:", err)
        exit()
    }
    // automatically call Close() at the end of current method
    defer file.Close()

    reader := csv.NewReader(file)

    raw, err := reader.ReadAll()
    if err != nil {
        fmt.Println(err)
        exit()
    }

    // populate doors property
    for i, door := range raw {
        // skip header
        if i == 0 {
            continue
        }

        name := door[0]
        input := door[1]
        output := door[2]

        fmt.Println("Checking DOOR #" + name, "...")

        out, err := exec.Command("sbox", device, "read-input", input).Output()
        if len(out) == 0 || err != nil {
            fmt.Println("Error! reading input")
            exit()
        }
        
        fmt.Print("INPUT : ", input, " => ", string(out))
        
        out, err = exec.Command("sbox", device, "read-coil", output).Output()
        if len(out) == 0 || err != nil {
            fmt.Println("Error! reading output")
            exit()
        }

        fmt.Print("OUTPUT: ", output, " => ", string(out))

        fmt.Println("---------------------------------------------------")
        
        if len(os.Args) == 3 && os.Args[2] == "sleep" {
            time.Sleep(1000 * time.Millisecond)
        }
    }
}

func exit() {
    fmt.Println("ERROR! exiting ...")
    os.Exit(1)
}
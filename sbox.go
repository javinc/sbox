package main

import (
    "bytes"
    "code.google.com/p/gopass"
    "encoding/json"
    "encoding/csv"
    "fmt"
    "github.com/goburrow/modbus"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "os/exec"
    "strconv"
    "strings"
)

type Extra struct {
    Name string
    Input string
}

type Response struct {
    Type string
    Msg string
}

type Door struct {
    Name string
    Output string
    Width string
    Height string
    Length string
}

type Config struct {
    Port int
    Path string
    Api struct {
        Ip string
        Port int
    }
    Smartbox struct {
        Name string
        Ip string
        Port int
    }
    Address struct {
        Street string
        City string
        State string
        Country string
        Postal string
        Area string
    }
    Deployer struct {
        Username string
        Password string
    }
    Doors []Door
    Extras []Extra
}

const (
    quantity = 9
    on = uint16(0xFF00)
    off = uint16(0x0000)

    sboxFile = "sbox"
    configFile = "./config.json"
    binPath = "/usr/bin/"
    deviceLabel = "Delta"
    deviceInterface = "/dev/ttyUSB0"

    api = "http://127.0.0.1:8888/register"
)

var (
    host string
    action string

    config Config
    response Response
    extra Extra
    unitUser string
    confirm string
    installOption string
)

func main() {
    // validate command
    if len(os.Args) <= 2 {
        if len(os.Args) == 2 {
            switch os.Args[1] {
            case "install":
                install()

                return
            case "install-interface":
                installOption = "interface"
                install()

                return
            }
        }

        help()

        return
    } else if len(os.Args) < 3 {
        fmt.Println("ERROR: host & action not specified")

        return
    }

    // set host
    host = os.Args[1]
    action = os.Args[2]

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
    execute(host, action, address)
}

func execute(host string, action string, address uint16) {
    // check if type of client
    // its tcp by default
    client := modbus.TCPClient(host)
    if string(host[0]) == "/" {
        client = modbus.ASCIIClient(host)        
    }

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

func install() {
    // confirm installation
    fmt.Println(`
--------------------------------------------------------------
                    SmartBox Unit Installer
--------------------------------------------------------------
do you want to continue? [Y/n]:`)
        fmt.Scanf("%v", &confirm)
        if(confirm != "Y") {
            exit()
        }

        // getting system user
        unitUser = os.Getenv("USER")

        // check setup files
        fmt.Println("checking setup files ...")
        checkSetupFiles([]string{
            "config.json",
            "sbox",
        })

        // copy sbox to bin
        fmt.Println("copying sbox to bin ...")
        _, err := exec.Command("sudo", "cp", sboxFile, binPath).Output()
        if err != nil {
            fmt.Println("You must be a sudoer!")
            exit()
        }

        // get device's vendor and product ID
        fmt.Println("creating device interface ...")
        createDeviceInterface()

        // check if interface created
        fmt.Println("checking interface created ...")
        out, err := exec.Command("ls", deviceInterface).Output()
        if len(out) == 0 || err != nil {
            exit()
        }

        // set owner to user
        fmt.Println("setting owner to " + unitUser + " user ...")
        _, err = exec.Command("sudo", "chown", unitUser + ":" + unitUser, deviceInterface).Output()
        if err != nil {
            fmt.Println(err)
            exit()
        }

        // check install option 
        if installOption == "interface" {
            fmt.Println("creating interface done")
            exit()
        }

        // get config json file
        fmt.Println("parsing config json file ...")
        getConfig()
        
        // check if deployer exist else throw invalid json file
        if len(config.Deployer.Username) == 0 {
            fmt.Println("invalid json file")
            exit()
        }

        // get config csv file
        fmt.Println("parsing config csv file ...")
        parseCsv(config.Path)

        // check smartbox doors if there's open
        fmt.Println("checking smartbox doors ...")
        if !checkDoors() {
            fmt.Println("some doors not calibrated properly")
            exit()  
        }

        // get deployer password
        password, err := gopass.GetPass("password for deployer " + config.Deployer.Username + ":")
        if err != nil {
            exit()
        }

        config.Deployer.Password = password

        fmt.Println("authenticating ...")
        register()

        fmt.Println("setup complete!")
    }

func checkDoors() bool {
    // check first door input
    out, err := exec.Command(sboxFile, deviceInterface, "read-input", config.Doors[0].Input).Output()
    if len(out) == 0 || strings.IndexAny(string(out), "1") > -1 || err != nil {
        return false
    }

    return true
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

        config.Doors = append(config.Doors, Door{
            door[0],
            door[1],
            door[2],
            door[3],
            door[4],
        })
    }
}

func register() {
    data, err := json.Marshal(config)
    if err != nil {
        fmt.Println(err)
        exit()
    }

    // normalize data
    raw := strings.ToLower(string(data))

    // send post json
    req, err := http.NewRequest("POST", api, bytes.NewBuffer([]byte(raw)))
    req.Header.Set("Content-Type", "application/json")

    fmt.Println("requesting ...")
    client := &http.Client{}
    r, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
        exit()
    }
    defer r.Body.Close()

    fmt.Println("response status is", r.Status)

    fmt.Println("parsing response body ...")
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Println(err)
        exit()
    }

    json.Unmarshal([]byte(string(body)), &response)
    // check if response register is good
    if response.Type == "success" {
        // show some info on valid response
        
        return
    }   

    // error
    fmt.Println(response.Msg)
    exit()
}

func createDeviceInterface() {
    raw := pipedGrepCmd("lsusb", deviceLabel)
    if len(raw) == 0 {
        fmt.Println("device not connected!")

        exit()
    }

    raw = strings.Split(raw, " ")[5]
    ids := strings.Split(raw, ":")

    _, err := exec.Command("modprobe", "usbserial", "vendor=0x" + ids[0], "product=0x" + ids[1]).Output()
    if err != nil {
        exit()
    }
}

func pipedGrepCmd(c1, c2 string) string {
    first := exec.Command(c1)
    second := exec.Command("grep", c2)

    reader, writer := io.Pipe()

    first.Stdout = writer
    second.Stdin = reader

    var buff bytes.Buffer
    second.Stdout = &buff

    first.Start()
    second.Start()
    first.Wait()
    writer.Close()
    second.Wait()

    return buff.String()
}

func checkSetupFiles(files []string) {
    for _, file := range files {
        _, err := ioutil.ReadFile(file)
        if err != nil {
            exit()
        }
    }
}

func getConfig() {
    file, err := ioutil.ReadFile(configFile)
    if err != nil {
        exit()
    }

    json.Unmarshal([]byte(file), &config)
}

func exit() {
    fmt.Println("ERROR! exiting ...")
    os.Exit(1)
}

func help() {
    fmt.Println(`
    Setup:
        sbox install        Setup
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
}
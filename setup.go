package sbox

import (
    "fmt"
    "bytes"
    "strings"
    "os"
    "os/exec"
    "encoding/json"
    "io"
    "io/ioutil"
    "code.google.com/p/gopass"
)

type Config struct {
    Port int
    Path string
    Address struct {
        Label string
        Street string
        City string
        State string
    }
    Deployer struct {
        Username string
        Password string
    }
}

const (
    sboxFile = "sbox"
    configFile = "./config.json"
    binPath = "/usr/bin/"
    deviceLabel = "Delta"
    deviceInterface = "/dev/ttyUSB0"
)

var (
    config Config
    unitUser string
    confirm string
)

func install() {
    // confirm installation
    fmt.Println(`
--------------------------------------------------------------
                    SmartBox Unit Installer
--------------------------------------------------------------
do you want to continue? [Y/n]: `)
    fmt.Scanf("%v", &confirm)
    if(confirm != "Y") {
        exit()
    }

    // getting system user
    unitUser = os.Getenv("USER")

    // check setup files
    fmt.Println("checking setup files...")
    checkSetupFiles()

    // copy sbox to bin
    fmt.Println("copying sbox to bin...")
    _, err := exec.Command("sudo", "cp", sboxFile, binPath).Output()
    if err != nil {
        fmt.Println("You must be a sudoer!")
        exit()
    }

    // get device's vendor and product ID
    fmt.Println("creating device interface...")
    createDeviceInterface()

    // check if interface created
    fmt.Println("checking interface created...")
    out, err := exec.Command("ls", deviceInterface).Output()
    if len(out) == 0 || err != nil {
        exit()
    }

    // set owner to user
    fmt.Println("setting owner to " + unitUser + " user...")
    _, err = exec.Command("sudo", "chown", unitUser + ":" + unitUser, deviceInterface).Output()
    if err != nil {
        fmt.Println(err)
        exit()
    }

    // get config json file
    fmt.Println("parsing config json file...")
    getConfig()

    // get deployer password
    password, err := gopass.GetPass("password for " + config.Deployer.Username + ": ")
    if err != nil {
        exit()
    }

    config.Deployer.Password = password

    fmt.Println("authenticating...")
    register()

    fmt.Println("setup complete!")
}

func register() {
    
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

func checkSetupFiles() {
    files := []string{
        "config.json",
        "sbox",
    }

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
    fmt.Println("ERROR! exiting...")
    os.Exit(1)
}
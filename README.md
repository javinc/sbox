# sbox
smartbox cli using modbus

## setup
- `go build sbox.go`
- `mv sbox /usr/bin`
- `sbox`

    Setup:
        sbox install            Setup
        sbox install-interface  Inteface only
    Usage:
        sbox [host] [action] [address]
    Action:
        read-coil               READ_COILS
        write-coil              WRITE_SINGLE_COILS
        read-input              READ_DESCRETE_INPUTS
    Sample:
        sbox /dev/ttyUSB0 read-coil 500
        sbox /dev/ttyUSB0 read-input 410
        sbox /dev/ttyUSB0 write-coil-on 500
        sbox /dev/ttyUSB0 write-coil-off 500

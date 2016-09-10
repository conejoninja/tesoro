How to test
===========
Go to the *tests* folder and run them with
```bash
// Put your device in bootloader mode
go test -v tesoro_bootloader_test.go
// Disconnect and connect your device in normal mode
go test -v tesoro_test.go
```

Running tests the *traditional* Go way (*go test*) will not work, as for tesoro_bootloader_test.go to run you need to put your device in *bootloader* mode, the rest of the tests are run in normal mode.


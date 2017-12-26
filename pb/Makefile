
TREZOR_COMMON=../trezor-common/protob/

check: config.pb messages.pb storage.pb types.pb

messages.pb: $(TREZOR_COMMON)/messages.proto
	protoc --proto_path=$(TREZOR_COMMON) --go_out=messages $<

types.pb: $(TREZOR_COMMON)/types.proto
	protoc --proto_path=$(TREZOR_COMMON) --go_out=types $<

config.pb: $(TREZOR_COMMON)/config.proto
	protoc --proto_path=$(TREZOR_COMMON) --go_out=. $<

storage.pb: $(TREZOR_COMMON)/storage.proto
	protoc --proto_path=$(TREZOR_COMMON) --go_out=. $<

clean:
	rm -f *.pb *.pb.go

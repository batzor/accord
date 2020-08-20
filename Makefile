gen:
	protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:pb --grpc-gateway_out=:pb

test:
	go test tests/basic_test.go

cert:
	cd cert; ./gen.sh; cd ..

.PHONY: cert



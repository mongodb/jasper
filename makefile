buildDir := build
srcFiles := $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -name "*_test.go" -not -path "*\#*")
testFiles := $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -path "*\#*")

compile:
	go build ./...
race:
	go test -count 1 -v -race ./...
test: 
	go test -v -cover ./...
coverage:$(buildDir)/cover.out
	@go tool cover -func=$< | sed -E 's%github.com/.*/jasper/%%' | column -t
coverage-html:$(buildDir)/cover.html

$(buildDir):$(srcFiles) compile
	@mkdir -p $@
$(buildDir)/cover.out:$(buildDir) $(testFiles) .FORCE
	go test -count 1 -v -coverprofile $@ -cover ./...
$(buildDir)/cover.html:$(buildDir)/cover.out
	go tool cover -html=$< -o $@
.FORCE:


proto:
	@mkdir -p jrpc/internal
	protoc --go_out=plugins=grpc:jrpc/internal *.proto
	@sed -i 's%context "golang.org/x/net/context"%"context"%g' jrpc/internal/jasper.pb.go
clean: 
	rm *.pb.go mv 

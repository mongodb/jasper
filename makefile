buildDir := build
srcFiles := $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -name "*_test.go" -not -path "*\#*")
testFiles := $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -path "*\#*")

_testPackages := ./ ./jrpc ./jrpc/internal

compile:
	go build $(_testPackages)
race:
	go test -count 1 -v -race $(_testPackages)
test: 
	go test -v -cover $(_testPackages)
coverage:$(buildDir)/cover.out
	@go tool cover -func=$< | sed -E 's%github.com/.*/jasper/%%' | column -t
coverage-html:$(buildDir)/cover.html

$(buildDir):$(srcFiles) compile
	@mkdir -p $@
$(buildDir)/cover.out:$(buildDir) $(testFiles) .FORCE
	go test -count 1 -v -coverprofile $@ -cover $(_testPackages)
$(buildDir)/cover.html:$(buildDir)/cover.out
	go tool cover -html=$< -o $@
.FORCE:


proto:
	@mkdir -p jrpc/internal
	protoc --go_out=plugins=grpc:jrpc/internal *.proto
	@sed -i 's%context "golang.org/x/net/context"%"context"%g' jrpc/internal/jasper.pb.go
clean: 
	rm *.pb.go mv 

vendor-clean:
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/evergreen-c/igimlet/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/mongodb/grip/vendor/golang.org/x/sys/
	rm -rf vendor/github.com/mongodb/grip/buildscripts/

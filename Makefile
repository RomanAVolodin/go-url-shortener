test-coverage:
	go test -v ./... -covermode=count -coverpkg=./... -coverprofile coverage.out
	go tool cover -html coverage.out -o coverage.html
	open coverage.html

run:
	SERVER_ADDRESS=localhost:8080 BASE_URL=http://localhost:8080 go run cmd/shortener/main.go

test-coverage-console:
	go test ./...  -coverpkg=./... -coverprofile coverage.out
	go tool cover -func coverage.out


test_bench:
	go test -bench=. ./...

profiles_diff:
	go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

get_profile_under_presure:
	go tool pprof -http=":9090" -seconds=30 http://localhost:8080/debug/pprof/heap > profile/result.pprof

get_info_from_profile:
	 go tool pprof -http=":9090" profile/base.pprof


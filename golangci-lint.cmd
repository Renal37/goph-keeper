docker run --rm -v %cd%:/app -w /app golangci/golangci-lint:v1.55.2 golangci-lint run -c .golangci.yml > ./golangci-lint/report-unformatted.json
docker run --rm -v %cd%:/app imega/jq -c . /app/golangci-lint/report-unformatted.json > ./golangci-lint/report.json
pause
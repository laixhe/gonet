module github.com/laixhe/gonet/imaging

go 1.25

replace github.com/laixhe/gonet/utils => ../utils

require (
	github.com/laixhe/gonet/utils v0.8.0
	golang.org/x/image v0.35.0
)

require golang.org/x/text v0.33.0 // indirect

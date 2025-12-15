module github.com/laixhe/gonet/imaging

go 1.25

replace github.com/laixhe/gonet/utils => ../utils

require (
	github.com/laixhe/gonet/utils v0.3.3
	golang.org/x/image v0.34.0
)

require golang.org/x/text v0.32.0 // indirect

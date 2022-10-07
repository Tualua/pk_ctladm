module github.com/Tualua/pk_ctladm

go 1.19

require (
	github.com/Tualua/pk_ctladm/pk_scst v0.0.0-20221006061759-514390471c0c
	github.com/akamensky/argparse v1.4.0
)

require (
	github.com/sirupsen/logrus v1.9.0
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace github.com/Tualua/pk_ctladm/pk_scst v0.0.0-20221006061759-514390471c0c => ./pk_scst

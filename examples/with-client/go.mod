module todo

go 1.23.1

replace go.trulyao.dev/robin => ../..

require (
	go.etcd.io/bbolt v1.3.11
	go.trulyao.dev/robin v0.0.0-00010101000000-000000000000
	go.trulyao.dev/seer v1.1.0
)

require (
	github.com/agnivade/levenshtein v1.2.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/matthewhartstonge/argon2 v1.0.1 // indirect
	go.trulyao.dev/mirror/v2 v2.7.1 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
)

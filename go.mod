module github.com/okex/adventure

go 1.16

require (
	github.com/Workiva/go-datastructures v1.0.53 // indirect
	github.com/btcsuite/btcd v0.22.1 // indirect
	github.com/ethereum/go-ethereum v1.10.8
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-kit/kit v0.12.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/okex/exchain v1.6.4
	github.com/okex/exchain-go-sdk v1.5.7-0.20230323100532-56d7a353728e
	github.com/onsi/gomega v1.19.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.2 // indirect
	github.com/petermattis/goid v0.0.0-20220712135657-ac599d9cba15
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.12.0
	github.com/status-im/keycard-go v0.0.0-20190424133014-d95853db0f48
	github.com/subosito/gotenv v1.4.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/net v0.0.0-20220617184016-355a448f1bc9 // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467 // indirect
	google.golang.org/grpc v1.47.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/ethereum/go-ethereum => github.com/okex/go-ethereum v1.10.8-oec2
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	//github.com/okex/exchain => github.com/okex/exchain v1.6.5
	github.com/tendermint/go-amino => github.com/okex/go-amino v0.15.1-okc4
//github.com/tendermint/tm-db => github.com/okex/exchain/t
)

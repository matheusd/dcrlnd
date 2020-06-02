module github.com/decred/dcrlnd

go 1.13

require (
	decred.org/dcrwallet v1.2.3-0.20200316140923-52363c4bcf4a
	git.schwanenlied.me/yawning/bsaes.git v0.0.0-20180720073208-c0276d75487e // indirect
	github.com/NebulousLabs/go-upnp v0.0.0-20181203152547-b32978b8ccbf
	github.com/Yawning/aez v0.0.0-20180408160647-ec7426b44926
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd v1.2.1-0.20191016204340-338ce9d7017d
	github.com/decred/dcrd/bech32 v1.0.0
	github.com/decred/dcrd/blockchain/stake v1.2.1
	github.com/decred/dcrd/blockchain/standalone v1.1.0
	github.com/decred/dcrd/blockchain/v2 v2.1.0
	github.com/decred/dcrd/chaincfg/chainhash v1.0.2
	github.com/decred/dcrd/chaincfg/v2 v2.3.0
	github.com/decred/dcrd/chaincfg/v3 v3.0.0-20200311044114-143c1884e4c8
	github.com/decred/dcrd/connmgr v1.1.0
	github.com/decred/dcrd/dcrec v1.0.0
	github.com/decred/dcrd/dcrec/secp256k1/v2 v2.0.0
	github.com/decred/dcrd/dcrjson/v2 v2.2.0
	github.com/decred/dcrd/dcrutil v1.4.0
	github.com/decred/dcrd/dcrutil/v2 v2.0.1
	github.com/decred/dcrd/dcrutil/v3 v3.0.0-20200311044114-143c1884e4c8
	github.com/decred/dcrd/hdkeychain/v2 v2.1.0
	github.com/decred/dcrd/hdkeychain/v3 v3.0.0-20200312163737-f8c59264b16d
	github.com/decred/dcrd/rpc/jsonrpc/types/v2 v2.0.0
	github.com/decred/dcrd/rpcclient/v5 v5.0.0
	github.com/decred/dcrd/txscript/v2 v2.1.0
	github.com/decred/dcrd/wire v1.3.0
	github.com/decred/lightning-onion/v3 v3.0.0
	github.com/decred/slog v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.8.6
	github.com/jackpal/gateway v1.0.5
	github.com/jackpal/go-nat-pmp v0.0.0-20170405195558-28a68d0c24ad
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/logrotate v1.0.0
	github.com/juju/clock v0.0.0-20180808021310-bab88fc67299 // indirect
	github.com/juju/errors v0.0.0-20181118221551-089d3ea4e4d5 // indirect
	github.com/juju/loggo v0.0.0-20180524022052-584905176618 // indirect
	github.com/juju/retry v0.0.0-20180821225755-9058e192b216 // indirect
	github.com/juju/testing v0.0.0-20180920084828-472a3e8b2073 // indirect
	github.com/juju/utils v0.0.0-20180820210520-bf9cc5bdd62d // indirect
	github.com/juju/version v0.0.0-20180108022336-b64dbd566305 // indirect
	github.com/kkdai/bstream v0.0.0-20181106074824-b3251f7901ec
	github.com/lightninglabs/protobuf-hex-display v1.3.3-0.20191212020323-b444784ce75d
	github.com/miekg/dns v1.1.3
	github.com/prometheus/client_golang v0.9.3
	github.com/tv42/zbase32 v0.0.0-20160707012821-501572607d02
	github.com/urfave/cli v1.20.0
	gitlab.com/NebulousLabs/fastrand v0.0.0-20181126182046-603482d69e40 // indirect
	gitlab.com/NebulousLabs/go-upnp v0.0.0-20181011194642-3a71999ed0d3 // indirect
	go.etcd.io/bbolt v1.3.4
	golang.org/x/crypto v0.0.0-20200214034016-1d94cc7ab1c6
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.27.1
	gopkg.in/errgo.v1 v1.0.0 // indirect
	gopkg.in/macaroon-bakery.v2 v2.1.0
	gopkg.in/macaroon.v2 v2.0.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
)

replace go.etcd.io/bbolt => github.com/matheusd/bbolt v1.3.5-0.20200328154133-ef6e10561575

replace github.com/decred/lightning-onion/v3 => github.com/matheusd/lightning-onion/v3 v3.0.0-20200108130659-ce91c24fe0c2

replace (
	decred.org/dcrwallet => github.com/matheusd/dcrwallet v1.0.9-0.20200417164917-5eb83e092aa8
	github.com/decred/dcrd/rpc/jsonrpc/types/v2 => github.com/matheusd/dcrd/rpc/jsonrpc/types/v2 v2.0.0-20200304154418-6667f0315aab
	github.com/decred/dcrd/rpcclient/v5 => github.com/matheusd/dcrd/rpcclient/v5 v5.0.0-20200321185658-73449c01a3eb
	github.com/jrick/wsrpc/v2 => github.com/jrick/wsrpc/v2 v2.3.3-0.20200304154051-c15991053cce
)

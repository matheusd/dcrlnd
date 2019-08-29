module github.com/decred/dcrlnd

require (
	git.schwanenlied.me/yawning/bsaes.git v0.0.0-20180720073208-c0276d75487e // indirect
	github.com/NebulousLabs/go-upnp v0.0.0-20181203152547-b32978b8ccbf
	github.com/Yawning/aez v0.0.0-20180408160647-ec7426b44926
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd v1.3.0
	github.com/decred/dcrd/bech32 v1.0.0
	github.com/decred/dcrd/blockchain v1.1.1
	github.com/decred/dcrd/blockchain/stake v1.1.0
	github.com/decred/dcrd/chaincfg v1.5.1
	github.com/decred/dcrd/chaincfg/chainhash v1.0.1
	github.com/decred/dcrd/connmgr v1.0.2
	github.com/decred/dcrd/dcrec v1.0.0
	github.com/decred/dcrd/dcrec/secp256k1 v1.0.2
	github.com/decred/dcrd/dcrjson v1.1.0 // indirect
	github.com/decred/dcrd/dcrjson/v2 v2.0.0
	github.com/decred/dcrd/dcrutil v1.3.0
	github.com/decred/dcrd/hdkeychain v1.1.1
	github.com/decred/dcrd/mempool/v2 v2.0.0
	github.com/decred/dcrd/rpcclient/v2 v2.0.0
	github.com/decred/dcrd/txscript v1.0.2
	github.com/decred/dcrd/wire v1.2.0
	github.com/decred/dcrwallet v1.2.2
	github.com/decred/dcrwallet/chain/v2 v2.0.0
	github.com/decred/dcrwallet/errors v1.0.1
	github.com/decred/dcrwallet/wallet/v2 v2.0.0
	github.com/decred/lightning-onion v1.0.0
	github.com/decred/slog v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/golang/protobuf v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.6.4
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
	github.com/miekg/dns v1.1.3
	github.com/prometheus/client_golang v0.9.3
	github.com/rogpeppe/fastuuid v0.0.0-20150106093220-6724a57986af // indirect
	github.com/tv42/zbase32 v0.0.0-20160707012821-501572607d02
	github.com/urfave/cli v1.20.0
	gitlab.com/NebulousLabs/fastrand v0.0.0-20181126182046-603482d69e40 // indirect
	gitlab.com/NebulousLabs/go-upnp v0.0.0-20181011194642-3a71999ed0d3 // indirect
	go.etcd.io/bbolt v1.3.3
	golang.org/x/crypto v0.0.0-20190611184440-5c40567a22f8
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c
	google.golang.org/genproto v0.0.0-20190111180523-db91494dd46c
	google.golang.org/grpc v1.18.0
	gopkg.in/errgo.v1 v1.0.0 // indirect
	gopkg.in/macaroon-bakery.v2 v2.1.0
	gopkg.in/macaroon.v2 v2.0.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
)

replace (
	github.com/decred/dcrd => github.com/decred/dcrd v0.0.0-20190410055418-5048959fdcf6
	github.com/decred/dcrd/bech32 => github.com/decred/dcrd/bech32 v0.0.0-20190410055418-5048959fdcf6
	github.com/decred/dcrd/blockchain => github.com/decred/dcrd/blockchain v0.0.0-20190410055418-5048959fdcf6
	github.com/decred/dcrd/connmgr => github.com/decred/dcrd/connmgr v0.0.0-20190410055418-5048959fdcf6
	github.com/decred/dcrd/dcrjson/v2 => github.com/decred/dcrd/dcrjson/v2 v2.0.0-20190410055418-5048959fdcf6
	github.com/decred/dcrd/mempool/v2 => github.com/decred/dcrd/mempool/v2 v2.0.0-20190524113024-5048959fdcf6
	github.com/decred/dcrd/peer => github.com/decred/dcrd/peer v0.0.0-20190410055418-5048959fdcf6
	github.com/decred/dcrd/rpcclient/v2 => github.com/decred/dcrd/rpcclient/v2 v2.0.0-20190410055418-5048959fdcf6
	github.com/decred/dcrd/txscript => github.com/decred/dcrd/txscript v0.0.0-20190410055418-5048959fdcf6

	github.com/decred/dcrwallet => github.com/decred/dcrwallet v0.0.0-20190322135901-7e0e5a4227d7
	github.com/decred/dcrwallet/wallet/v2 => github.com/decred/dcrwallet/wallet/v2 v2.0.0-20190322135901-7e0e5a4227d7

	github.com/decred/lightning-onion => github.com/matheusd/lightning-onion v0.0.0-20190723174454-28be9117331b
)

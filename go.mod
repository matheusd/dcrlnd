module github.com/decred/dcrlnd

go 1.13

require (
	decred.org/dcrwallet v1.2.3-0.20200618173758-53994fa373a2
	git.schwanenlied.me/yawning/bsaes.git v0.0.0-20180720073208-c0276d75487e // indirect
	github.com/NebulousLabs/go-upnp v0.0.0-20181203152547-b32978b8ccbf
	github.com/Yawning/aez v0.0.0-20180408160647-ec7426b44926
	github.com/btcsuite/btcwallet/walletdb v1.3.3
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd v1.2.1-0.20200604211334-3f43437d1338
	github.com/decred/dcrd/bech32 v1.0.0
	github.com/decred/dcrd/blockchain/stake/v3 v3.0.0-20200608124004-b2f67c2dc475
	github.com/decred/dcrd/blockchain/standalone/v2 v2.0.0-20200721173351-66807231ce05
	github.com/decred/dcrd/blockchain/v3 v3.0.0-20200608124004-b2f67c2dc475
	github.com/decred/dcrd/chaincfg/chainhash v1.0.2
	github.com/decred/dcrd/chaincfg/v3 v3.0.0-20200608124004-b2f67c2dc475
	github.com/decred/dcrd/connmgr v1.1.0
	github.com/decred/dcrd/dcrec v1.0.0
	github.com/decred/dcrd/dcrec/secp256k1/v2 v2.0.0
	github.com/decred/dcrd/dcrec/secp256k1/v3 v3.0.0
	github.com/decred/dcrd/dcrjson/v2 v2.2.0
	github.com/decred/dcrd/dcrutil/v3 v3.0.0-20200608124004-b2f67c2dc475
	github.com/decred/dcrd/gcs/v2 v2.0.2-0.20200608124004-b2f67c2dc475
	github.com/decred/dcrd/hdkeychain/v3 v3.0.0
	github.com/decred/dcrd/rpc/jsonrpc/types/v2 v2.0.1-0.20200503044000-76f6906e50e5
	github.com/decred/dcrd/rpcclient/v6 v6.0.0-20200616182840-3baf1f590cb1
	github.com/decred/dcrd/txscript/v3 v3.0.0-20200608124004-b2f67c2dc475
	github.com/decred/dcrd/wire v1.3.0
	github.com/decred/lightning-onion/v3 v3.0.0-20200907144804-1e4209f3f75a
	github.com/decred/slog v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/strfmt v0.19.5 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.3
	github.com/jackpal/gateway v1.0.5
	github.com/jackpal/go-nat-pmp v0.0.0-20170405195558-28a68d0c24ad
	github.com/jedib0t/go-pretty v4.3.0+incompatible
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
	github.com/matheusd/etcd v0.5.0-alpha.5.0.20201012144914-e63dcc3d7528
	github.com/matheusd/google-protobuf-protos v0.0.0-20200707194502-ef6ec5c2266f
	github.com/matheusd/protobuf-hex-display v1.3.3-0.20201012153224-75fb8d4840f1
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/miekg/dns v1.1.3
	github.com/prometheus/client_golang v0.9.3
	github.com/stretchr/testify v1.4.0
	github.com/tv42/zbase32 v0.0.0-20160707012821-501572607d02
	github.com/urfave/cli v1.20.0
	gitlab.com/NebulousLabs/fastrand v0.0.0-20181126182046-603482d69e40 // indirect
	gitlab.com/NebulousLabs/go-upnp v0.0.0-20181011194642-3a71999ed0d3 // indirect
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/genproto v0.0.0-20200806141610-86f49bd18e98
	google.golang.org/grpc v1.31.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/errgo.v1 v1.0.0 // indirect
	gopkg.in/macaroon-bakery.v2 v2.1.0
	gopkg.in/macaroon.v2 v2.0.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
)

replace (
	github.com/decred/dcrd/dcrutil/v3 => github.com/decred/dcrd/dcrutil/v3 v3.0.0-20200604211334-3f43437d1338
	github.com/decred/dcrd/hdkeychain/v3 => github.com/decred/dcrd/hdkeychain/v3 v3.0.0-20200604211334-3f43437d1338
	github.com/decred/dcrd/txscript/v3 => github.com/decred/dcrd/txscript/v3 v3.0.0-20200604211334-3f43437d1338
)

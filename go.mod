module github.com/decred/dcrlnd

require (
	git.schwanenlied.me/yawning/bsaes.git v0.0.0-20180720073208-c0276d75487e // indirect
	github.com/NebulousLabs/go-upnp v0.0.0-20181203152547-b32978b8ccbf
	github.com/Roasbeef/btcutil v0.0.0-20180406014609-dfb640c57141 // indirect
	github.com/Yawning/aez v0.0.0-20180408160647-ec7426b44926
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da // indirect
	github.com/awalterschulze/gographviz v0.0.0-20181013152038-b2885df04310
	github.com/btcsuite/btcd v0.0.0-20190109040709-5bda5314ca95
	github.com/btcsuite/btcutil v0.0.0-20190112041146-bf1e1be93589
	github.com/coreos/bbolt v1.3.0
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd v1.3.0
	github.com/decred/dcrd/blockchain v1.1.1
	github.com/decred/dcrd/blockchain/stake v1.1.0
	github.com/decred/dcrd/certgen v1.0.2
	github.com/decred/dcrd/chaincfg v1.3.0
	github.com/decred/dcrd/chaincfg/chainhash v1.0.1
	github.com/decred/dcrd/connmgr v1.0.2
	github.com/decred/dcrd/dcrec v0.0.0-20190214012338-9265b4051009
	github.com/decred/dcrd/dcrec/secp256k1 v1.0.1
	github.com/decred/dcrd/dcrjson/v2 v2.0.0
	github.com/decred/dcrd/dcrutil v1.2.0
	github.com/decred/dcrd/hdkeychain v1.1.1
	github.com/decred/dcrd/mempool v1.1.1
	github.com/decred/dcrd/rpcclient/v2 v2.0.0
	github.com/decred/dcrd/txscript v1.0.2
	github.com/decred/dcrd/wire v1.2.0
	github.com/decred/dcrwallet v1.2.2
	github.com/decred/dcrwallet/chain/v2 v2.0.0
	github.com/decred/dcrwallet/errors v1.0.1
	github.com/decred/dcrwallet/ticketbuyer v1.0.2 // indirect
	github.com/decred/dcrwallet/wallet/v2 v2.0.0
	github.com/decred/lnd v0.2.1-alpha
	github.com/decred/slog v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/golang/protobuf v1.2.0
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
	github.com/lightningnetwork/lightning-onion v0.0.0-20181120012844-725fcf830eac
	github.com/lightningnetwork/lnd v0.0.2 // indirect
	github.com/ltcsuite/ltcd v0.0.0-20190101042124-f37f8bf35796
	github.com/miekg/dns v1.1.3
	github.com/roasbeef/btcd v0.0.0-20180418012700-a03db407e40d // indirect
	github.com/roasbeef/btcrpcclient v0.0.0-20170622074026-d0f4db8b4dad // indirect
	github.com/roasbeef/btcutil v0.0.0-20180406014609-dfb640c57141
	github.com/rogpeppe/fastuuid v0.0.0-20150106093220-6724a57986af // indirect
	github.com/tv42/zbase32 v0.0.0-20160707012821-501572607d02
	github.com/urfave/cli v1.20.0
	gitlab.com/NebulousLabs/fastrand v0.0.0-20181126182046-603482d69e40 // indirect
	gitlab.com/NebulousLabs/go-upnp v0.0.0-20181011194642-3a71999ed0d3 // indirect
	go.etcd.io/bbolt v1.3.2
	golang.org/x/crypto v0.0.0-20190103213133-ff983b9c42bc
	golang.org/x/net v0.0.0-20190110200230-915654e7eabc
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c
	google.golang.org/genproto v0.0.0-20190111180523-db91494dd46c
	google.golang.org/grpc v1.18.0
	gopkg.in/errgo.v1 v1.0.0 // indirect
	gopkg.in/macaroon-bakery.v2 v2.1.0
	gopkg.in/macaroon.v1 v1.0.0-20170816141150-ab101776739e
	gopkg.in/macaroon.v2 v2.0.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
)

replace (
	github.com/decred/dcrd => github.com/matheusd/dcrd v0.0.0-20190306151227-fe7df03d5792
	github.com/decred/dcrd/connmgr => github.com/matheusd/dcrd/connmgr v0.0.0-20190306151227-fe7df03d5792
	github.com/decred/dcrd/rpctest => github.com/davecgh/btcd/rpctest v0.0.0-20190306151227-fe7df03d5792

	github.com/decred/dcrwallet => github.com/matheusd/dcrwallet v0.0.0-20190225211924-3a618c61edb8
	github.com/decred/dcrwallet/wallet/v2 => github.com/matheusd/dcrwallet/wallet/v2 v2.0.0-20190225211924-3a618c61edb8
)

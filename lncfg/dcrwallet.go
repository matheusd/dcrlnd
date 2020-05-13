package lncfg

type DcrwalletConfig struct {
	GRPCHost      string `long:"grpchost" description:"The wallet's grpc listening address. If a port is omitted, then the default port for the selected chain parameters will be used."`
	CertPath      string `long:"certpath" description:"The file containing the wallet's certificate file."`
	AccountNumber int32  `long:"accountnumber" description:"The account number that dcrlnd should take control of for all onchain operations and offchain key derivation."`
}


geth init --datadir  <Datadir> genesis.json

geth --datadir  <Datadir>  --networkid <networkid> --nodiscover --http --rpc --rpcport "8545" --rpcaddr "0.0.0.0" --rpccorsdomain "*" --rpcapi "eth,web3,net,personal,miner" console 2
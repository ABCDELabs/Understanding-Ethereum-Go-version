# 1. First init new genesis state
geth init --datadir  <Datadir> genesis.json

# 2. Based on Local Chaindata and the Newwork id, start Geth.
geth --datadir  <Datadir>  --networkid <networkid> --nodiscover --http --rpc --rpcport "8545" --rpcaddr "0.0.0.0" --rpccorsdomain "*" --rpcapi "eth,web3,net,personal,miner" console 2


